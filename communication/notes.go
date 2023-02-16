package communication

import (
	notesv1 "accounts-service/protorepo/noted/notes/v1"

	"google.golang.org/grpc"
)

type NoteServiceClient struct {
	conn   *grpc.ClientConn
	Notes  notesv1.NotesAPIClient
	Groups notesv1.GroupsAPIClient
}

func NewNoteServiceClient(address string) (*NoteServiceClient, error) {
	res := NoteServiceClient{}

	err := res.Init(address)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *NoteServiceClient) Init(address string) error {
	conn, err := grpc.Dial(address)
	if err != nil {
		return err
	}

	c.conn = conn
	c.Notes = notesv1.NewNotesAPIClient(c.conn)
	c.Groups = notesv1.NewGroupsAPIClient(c.conn)

	return nil
}

func (c *NoteServiceClient) Close(address string) error {
	return c.conn.Close()
}
