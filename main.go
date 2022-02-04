package main

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"cloud.google.com/go/storage"
	"github.com/fsnotify/fsnotify"
)

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// config options
	watchDir := "./"
	pcapFileNamePrefix := "./dump.pcap"
	bucketName := "network-monitor-storage"
	storageFileNamePrefix := "dump"

	watchFile := ""
	initializedWatchFile := false

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Error("error:", err)
		return
	}

	bkt := client.Bucket(bucketName)

	done := make(chan bool)
	go func() {
		defer func() { done <- true }() // make sure to close the program in case the channels are closed
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					// log.WithField("filename", event.Name).Info("file write")
					// file was modified
					if strings.HasPrefix(event.Name, pcapFileNamePrefix) {
						// only handle pcap files
						if !initializedWatchFile {
							// watch file has not been initialized, this is the start point
							watchFile = event.Name
							initializedWatchFile = true
						} else {
							// watch file has already been initialized, check if modified file is different
							if event.Name != watchFile {
								// modified file is different, archive watchFile
								// get writer to storageObj
								storageFileName := storageFileNamePrefix + event.Name[len(event.Name)-2:] + "-" + time.Now().UTC().Format("200601021504")
								logger := log.WithField("filename", storageFileName)

								storageObj := bkt.Object(storageFileName)
								storageObjWriter := storageObj.NewWriter(ctx)

								// get reader of file to archive
								archiveFile, err := os.Open(watchFile)
								if err != nil {
									logger.Error("read archive file error:", err)
									panic(err)
								}

								// copy bytes from reader to writer
								bytes, err := io.Copy(storageObjWriter, archiveFile)
								if err != nil {
									logger.Error("copy error:", err)
									panic(err)
								}

								if err := storageObjWriter.Close(); err != nil {
									logger.Error("close error:", err)
									panic(err)
								}

								logger.WithField("bytes", bytes).Info("uploaded archive")

								// update the watchFile for when the next file is modified in the logrotate
								watchFile = event.Name
							}
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error("watcher error:", err)
			}
		}
	}()

	err = watcher.Add(watchDir)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Pcap Uploader is active!")
	<-done
}
