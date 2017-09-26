# Google Drive Block Device

This is a POC for using gdrive as a backend for a block-device.

## Roadmap

* [x] Memory backed chunked device
* [ ] Disk backed chunked device
* [ ] Gdrive backed chunked device
* [ ] Memory caching for gdrive chunks
* [ ] Disk chunk caching
* [ ] Optimize read/writes

## Usage

```
go run *.go
mkfs.ext4 -O ^has_journal /dev/nbd0
mkdir /mnt/mem
mount /dev/nbd0 /mnt/mem
echo "Hello world" > /mnt/mem/hello.txt
cat /mnt/mem/README.md
ls /mnt/mem
```