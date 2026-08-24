package main

import ( 
	"context"
	"time"


	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Message struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Content   string    `bson:"content" json:"content"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

type Store struct{
	collection *mongo.Collection
}

func newStore(ctx context.Context, uri string) (*Store, error){
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	collection := client.Database("chatdb").Collection("messages")
	return &Store{collection: collection}, nil
}

func (s *Store) SaveMessage(ctx context.Context, content string) (Message, error){
	msg := Message {
		Content: content,
		CreatedAt: time.Now(),
	}

	result, err := s.collection.InsertOne(ctx, msg)
	if err != nil {
		return msg, err
	}

	if oid, ok := result.InsertedID.(interface{ Hex() string }); ok {
		msg.ID = oid.Hex()
	}
	return msg, nil
}


func (s *Store) RecentMessages(ctx context.Context, limit int64) ([]Message, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit)
	cursor, err := s.collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []Message

	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}


	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}