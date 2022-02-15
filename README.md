# pcap_uploader

This repository contains the code for the pcap uploader that uploads pcaps generated from tcpdump on rotation to a GCS bucket.

The scripts folders contains the scripts to both start the tcpdump process as well as the uploader process. It also includes service files used to make the programs persist even on restart.

The pcap uploader program itself is written in golang and uses fsnotify as well as google's api to upload to GCS. Essentially it watches for file changes and if the new file changed is different from the previous file it was watching then it uploads the older file to GCS. This is because we know tcpdump has rotated to a new file and begun writting to it. It also ensures that we never lose data because all files in the watched directory are guaranteed to be uploaded. Although short the idea is pretty neat and thanks to aspyhxia for introducing me to initofy interrupts and fsnotify!
