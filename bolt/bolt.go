// Package bdb provides a FIFO queue backed by a BoltDB database.
package bdb

import (
	"encoding/binary"
	"time"

	"github.com/boltdb/bolt"
	pb "github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/negz/q"
	"github.com/negz/q/e"
	"github.com/negz/q/proto"
)

var (
	keyMetadata = []byte("meta")
	keyLimit    = []byte("limit")
	keyMessages = []byte("messages")
)

type bdb struct {
	meta  *q.Metadata
	limit int
	db    *bolt.DB
}

// An Option represents an optional argument to a new BoltDB queue.
type Option func(*bdb)

// Limit specifies the maximum number of messages that may exist in a queue.
// Unbounded queues will accept messages until they exhaust available resources.
func Limit(l int) Option {
	return func(b *bdb) {
		b.limit = l
	}
}

// Tagged applies the provided tags to a new queue.
func Tagged(t ...q.Tag) Option {
	return func(b *bdb) {
		for _, tag := range t {
			b.meta.Tags.AddTag(tag)
		}
	}
}

// New creates a new BoltDB backed FIFO queue.
func New(db *bolt.DB, o ...Option) (q.Queue, error) {
	id := uuid.New()
	meta := &q.Metadata{ID: id, Created: time.Now(), Tags: &q.Tags{}}
	queue := &bdb{meta: meta, limit: q.Unbounded, db: db}
	for _, opt := range o {
		opt(queue)
	}

	pmeta, err := proto.FromMeta(meta)
	if err != nil {
		return nil, errors.Wrap(err, "cannot marshal metadata to protobuf")
	}
	bmeta, err := pb.Marshal(pmeta)
	if err != nil {
		return nil, errors.Wrap(err, "cannot marshal metadata to bytes")
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		// uuid.UUID is a 16 byte array. id[:] converts it to a byte slice.
		bucket, err := tx.CreateBucket(id[:])
		if err != nil {
			return errors.Wrapf(err, "cannot create new BoltDB bucket %s", id)
		}

		if err := bucket.Put(keyMetadata, bmeta); err != nil {
			return errors.Wrap(err, "cannot store metadata")
		}
		if err := bucket.Put(keyLimit, itob(queue.limit)); err != nil {
			return errors.Wrap(err, "cannot store limit")
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "cannot store queue in BoltDB")
	}
	return queue, nil
}

// Open an existing BoltDB backed FIFO queue.
func Open(db *bolt.DB, id uuid.UUID) (q.Queue, error) {
	queue := &bdb{meta: &q.Metadata{}, limit: q.Unbounded, db: db}
	if err := db.View(func(tx *bolt.Tx) error {
		// uuid.UUID is a 16 byte array. id[:] converts it to a byte slice.
		bucket := tx.Bucket(id[:])
		if bucket == nil {
			return e.ErrNotFound(errors.Errorf("cannot open bucket %s", id))
		}

		bmeta := bucket.Get(keyMetadata)
		if bmeta == nil {
			return errors.New("cannot read queue metadata")
		}
		pmeta := &proto.Metadata{}
		if err := pb.Unmarshal(bmeta, pmeta); err != nil {
			return errors.Wrap(err, "cannot unmarshal queue metadata from bytes to protobuf")
		}
		meta, err := proto.ToMeta(pmeta)
		if err != nil {
			return errors.Wrap(err, "cannot convert metadata from protobuf")
		}
		queue.meta = meta

		blimit := bucket.Get(keyLimit)
		if blimit == nil {
			return errors.New("cannot read queue limit")
		}
		queue.limit = btoi(blimit)

		return nil
	}); err != nil {
		return nil, errors.Wrapf(err, "cannot read queue %s from BoltDB", id)
	}
	return queue, nil
}

func itob(i int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(i))
	return b
}

func btoi(b []byte) int {
	return int(binary.BigEndian.Uint64(b))
}

func (b *bdb) ID() uuid.UUID {
	return b.meta.ID
}

func (b *bdb) Store() q.Store {
	return q.BoltDB
}

func (b *bdb) Created() time.Time {
	return b.meta.Created
}

// TODO(negz): BoltDb implementation of tags too. :(
func (b *bdb) Tags() *q.Tags {
	return b.meta.Tags
}

// We key messages in the messages bucket using the bucket's monotonically
// increasing NextSequence method. Thus the length of the queue should be the
// key of the last message minus the key of the first message.
func getLength(b *bolt.Bucket) int {
	msgs := b.Bucket(keyMessages)
	if msgs == nil {
		return 0
	}
	c := msgs.Cursor()
	f, _ := c.First()
	if f == nil {
		return 0
	}
	l, _ := c.Last()
	return (btoi(l) - btoi(f)) + 1
}

func (b *bdb) Add(m *q.Message) error {
	pmsg, err := proto.FromMessage(m)
	if err != nil {
		return errors.Wrap(err, "cannot marshal message to protobuf")
	}
	bmsg, err := pb.Marshal(pmsg)
	if err != nil {
		return errors.Wrap(err, "cannot marshal message to bytes")
	}
	err = b.db.Update(func(tx *bolt.Tx) error {
		id := b.ID()
		bucket := tx.Bucket(id[:])
		if bucket == nil {
			return e.ErrNotFound(errors.Errorf("cannot open BoltDB bucket %s", b.ID()))
		}

		length := getLength(bucket)
		if (b.limit != q.Unbounded) && (length >= b.limit) {
			return e.ErrFull(errors.Errorf("queue %s has reached limit of %d messages", b.ID(), b.limit))
		}

		msgs, berr := bucket.CreateBucketIfNotExists(keyMessages)
		if berr != nil {
			return errors.Wrap(berr, "cannot create messages bucket")
		}

		// This returns an error only if the Tx is closed or not writeable,
		// which can't happen inside an update.
		i, _ := msgs.NextSequence()
		if perr := msgs.Put(itob(int(i)), bmsg); perr != nil {
			return errors.Wrap(perr, "cannot store message")
		}
		return nil
	})
	return errors.Wrap(err, "cannot store message in queue")
}

func (b *bdb) Pop() (*q.Message, error) {
	var msg *q.Message
	if err := b.db.Update(func(tx *bolt.Tx) error {
		id := b.ID()
		bucket := tx.Bucket(id[:])
		if bucket == nil {
			return e.ErrNotFound(errors.Errorf("cannot open BoltDB bucket %s", b.ID()))
		}

		msgs := bucket.Bucket(keyMessages)
		if msgs == nil {
			return e.ErrNotFound(errors.Errorf("queue %s is empty", b.ID()))
		}
		k, bmsg := msgs.Cursor().First()
		if k == nil {
			return e.ErrNotFound(errors.Errorf("queue %s is empty", b.ID()))
		}
		if err := msgs.Delete(k); err != nil {
			return errors.Wrap(err, "cannot delete message")
		}
		pmsg := &proto.Message{}
		if err := pb.Unmarshal(bmsg, pmsg); err != nil {
			return errors.Wrap(err, "cannot unmarshal message from bytes to protobuf")
		}
		var err error
		msg, err = proto.ToMessage(pmsg)
		if err != nil {
			return errors.Wrap(err, "cannot convert message from protobuf")
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "cannot pop from queue")
	}
	return msg, nil
}

func (b *bdb) Peek() (*q.Message, error) {
	var msg *q.Message
	if err := b.db.View(func(tx *bolt.Tx) error {
		id := b.ID()
		bucket := tx.Bucket(id[:])
		if bucket == nil {
			return e.ErrNotFound(errors.Errorf("cannot open BoltDB bucket %s", b.ID()))
		}
		msgs := bucket.Bucket(keyMessages)
		if msgs == nil {
			return e.ErrNotFound(errors.Errorf("queue %s is empty", b.ID()))
		}
		k, bmsg := msgs.Cursor().First()
		if k == nil {
			return e.ErrNotFound(errors.Errorf("queue %s is empty", b.ID()))
		}
		pmsg := &proto.Message{}
		if err := pb.Unmarshal(bmsg, pmsg); err != nil {
			return errors.Wrap(err, "cannot unmarshal message from bytes to protobuf")
		}
		var err error
		msg, err = proto.ToMessage(pmsg)
		if err != nil {
			return errors.Wrap(err, "cannot convert message from protobuf")
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "cannot peek into queue")
	}
	return msg, nil
}
