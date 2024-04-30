// package datastore

// import (
// 	"context"
// 	"log"
// 	"time"

// 	clientv3 "go.etcd.io/etcd/client/v3"
// )

// type DataStore struct {
// 	Config clientv3.Config
// 	Client *clientv3.Client
// }

// func (d *DataStore) Init(endpoints []string, dialtimeout time.Duration) *DataStore {
// 	// Configure Etcd3 client
// 	d.Config = clientv3.Config{
// 		Endpoints:   endpoints,
// 		DialTimeout: dialtimeout,
// 	}
// 	// Connect with config
// 	cli, err := clientv3.New(d.Config)
// 	if err != nil {
// 		panic(err)
// 	}
// 	// Set client connection
// 	d.Client = cli
// 	return d
// }

// func (d *DataStore) ResetGame() error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
// 	// Delete whole datastore
// 	_, err := d.Client.Delete(ctx, "", clientv3.WithPrefix(), clientv3.WithPrevKV())
// 	cancel()
// 	if err != nil {
// 		log.Panicf("Could not delete key %v", err)
// 		return err
// 	}
// 	log.Println("Deleted the whole state.")
// 	return nil
// }
