package service

import (
	"context"

	"cloud.google.com/go/firestore"
)

type (
	Datastore struct {
		ProjectID string
	}
)

func NewDatastore(projectID string) *Datastore {
	return &Datastore{ProjectID: projectID}
}

func (d *Datastore) Get(ctx context.Context, path string, dest interface{}) error {

	// if os.Getenv("GAE_APPLICATION") == "" {
	// 	if strings.HasPrefix(path, "Teams/") {
	// 		reflect.ValueOf(dest).Elem().
	// 			FieldByName("AccessToken").
	// 			Set(reflect.ValueOf(os.Getenv("SLACK_BOT_USER_OAUTH_ACCESS_TOKEN")))
	// 	}
	// 	return nil
	// }

	c, err := firestore.NewClient(ctx, d.ProjectID)
	if err != nil {
		return err
	}
	defer c.Close()

	doc, err := c.Doc(path).Get(ctx)
	if err != nil {
		return err
	}

	if err := doc.DataTo(dest); err != nil {
		return err
	}

	return nil
}

func (d *Datastore) Set(ctx context.Context, path string, val interface{}) error {

	// if os.Getenv("GAE_APPLICATION") == "" { // Do nothing on local
	// 	return nil
	// }

	c, err := firestore.NewClient(ctx, d.ProjectID)
	if err != nil {
		return err
	}
	defer c.Close()

	if _, err := c.Doc(path).Set(ctx, val); err != nil {
		return err
	}
	return nil
}
