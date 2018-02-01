package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/kelseyhightower/envconfig"
	"github.com/mholt/archiver"
	"google.golang.org/api/option"
)

const (
	envConfigPrefix = "pruls"
	filenameSep     = "_"
	filenameSuffix  = "backup"
	fileExtension   = ".tar.gz"
)

//Specification struct for keylseyhightower's envonfigs
type Specification struct {
	AccountKeyAbsPath string `default:"/etc/secret-volume/auth.json"`
	AppName           string `required:"true"`
	BucketName        string `required:"true"`
	FilePrefix        string `default:""`
	TargetDirAbsPath  string `required:"true"`
}

func main() {
	var s Specification
	err := envconfig.Process(envConfigPrefix, &s)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = validateEnvConfig(s)
	if err != nil {
		log.Fatal(err.Error())
	}
	ctx := context.Background()
	archiveFilename := archiveFilename(s.FilePrefix, s.AppName)
	obj := bucketObject(ctx, s.AccountKeyAbsPath, s.BucketName, archiveFilename)
	w := obj.NewWriter(ctx)
	err = archiver.TarGz.Make(archiveFilename, []string{s.TargetDirAbsPath})
	// open input file
	fi, err := os.Open(archiveFilename)
	if err != nil {
		log.Fatal(err.Error())
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err = fi.Close(); err != nil {
			log.Fatal(err.Error())
		}
	}()
	writeChunks(fi, w)
	log.Println("chunks written to " + archiveFilename + " in google bucket")
	verifyFileInBucket(ctx, obj)
	removeLocalFile(archiveFilename)

}

//removeLocalFile deletes the file with the specified filename
//It prints either a success or fail message to log
func removeLocalFile(archiveFilename string) {
	err := os.Remove(archiveFilename)
	if err == nil {
		log.Println("local file: " + archiveFilename + " has been deleted")
	} else {
		log.Println("problem deleting local file: ", err.Error())
	}
}

//verifyFileInBucket creates a bucket reader from the context and ObjectHandle
//If there's an error doing so, it calls log.Fatal
func verifyFileInBucket(ctx context.Context, obj storage.ObjectHandle) {
	bktReader, err := obj.NewReader(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	bktReader.Close()
}

//writeChunks reads chunks from the file and writes them to the storage.Writer
func writeChunks(fi *os.File, w *storage.Writer) {
	// make a read buffer
	fileReader := bufio.NewReader(fi)
	buf := make([]byte, 1024)
	for {
		// read a chunk
		n, err := fileReader.Read(buf)
		if err != nil && err != io.EOF {
			log.Fatal(err.Error())
		}
		if n == 0 {
			break
		}
		// write a chunk
		if _, err := w.Write(buf[:n]); err != nil {
			log.Fatal(err.Error())
		}
	}
	// Close, just like writing a file.
	if err := w.Close(); err != nil {
		log.Fatal(err.Error())
	}
}

//archiveFilename creates a filename from current time, file prefix (if exists),
// appname, suffix and fileExtension (in that order)
func archiveFilename(filePrefix, appName string) (filename string) {
	var buffer bytes.Buffer
	var timeFormatBuff bytes.Buffer
	timeFormatBuff.WriteString("2006")
	timeFormatBuff.WriteString(filenameSep)
	timeFormatBuff.WriteString("01")
	timeFormatBuff.WriteString(filenameSep)
	timeFormatBuff.WriteString("02")
	timeFormatBuff.WriteString(filenameSep)
	timeFormatBuff.WriteString("1504")
	timeFormatBuff.WriteString(filenameSep)
	timeFormatBuff.WriteString("05")
	buffer.WriteString(time.Now().Format(timeFormatBuff.String()))
	buffer.WriteString(filenameSep)
	if filePrefix != "" {
		buffer.WriteString(filePrefix)
		buffer.WriteString(filenameSep)
	}
	buffer.WriteString(appName)
	buffer.WriteString(filenameSep)
	buffer.WriteString(filenameSuffix)
	buffer.WriteString(fileExtension)
	filename = buffer.String()
	return
}

//bucketObject creates a storage.ObjectHandle from context, account key,
// bucket name and local file name
// It's here that the key (preferably a service-account .json) is used to auth
func bucketObject(ctx context.Context, accountKeyAbsPath, bucketName,
	archiveFilename string) (obj storage.ObjectHandle) {
	client, err := storage.NewClient(ctx,
		option.WithServiceAccountFile(accountKeyAbsPath))
	if err != nil {
		log.Fatal(err.Error())
	}
	bkt := client.Bucket(bucketName)
	obj = *bkt.Object(archiveFilename)
	return
}

//validateEnvConfig performs validation on envconfig
func validateEnvConfig(s Specification) (err error) {
	switch {
	case !pathStatBool(s.AccountKeyAbsPath):
		err = errors.New("google account key (e.g. .json) must exist, not found " +
			"at: " + s.AccountKeyAbsPath)
		break
	case !pathStatBool(s.TargetDirAbsPath):
		err = errors.New("backup target dir must exist, not found at: " +
			s.TargetDirAbsPath)
		break
	}
	return
}

//pathStatBool returns true if the path exists
func pathStatBool(path string) (exists bool) {
	_, err := os.Stat(path)
	exists = err == nil
	return
}
