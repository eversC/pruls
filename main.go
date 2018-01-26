package main

import (
	"bufio"
	"context"
	"io"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

func main() {

	ctx := context.Background()
	client, err := storage.NewClient(ctx,
		option.WithServiceAccountFile("/path/to/service-account-key.json"))
	if err != nil {
		// TODO: Handle error.
	}
	//TODO: bucket name from env var?
	bkt := client.Bucket("")
	obj := bkt.Object("data")
	w := obj.NewWriter(ctx)

  //TODO: generate filename from version (env var?) + timestamp

	//TODO: use something like https://github.com/mholt/archiver (with dir from env var?) to create tar.gz

	// open input file
	fi, err := os.Open("backup.tar.gz")
	if err != nil {
		// TODO: Handle error.
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err = fi.Close(); err != nil {
			// TODO: Handle error.
		}
	}()
	// make a read buffer
	fileReader := bufio.NewReader(fi)

	buf := make([]byte, 1024)

	for {
		// read a chunk
		n, err := fileReader.Read(buf)
		if err != nil && err != io.EOF {
			// TODO: Handle error.
		}
		if n == 0 {
			break
		}

		// write a chunk
		if _, err := w.Write(buf[:n]); err != nil {
			// TODO: Handle error.
		}
	}
	// Close, just like writing a file.
	if err = w.Close(); err != nil {
		// TODO: Handle error.
	}

	// Read it back.
	bktReader, err := obj.NewReader(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	defer bktReader.Close()
	if _, err := io.Copy(os.Stdout, bktReader); err != nil {
		// TODO: Handle error.
	}
}
