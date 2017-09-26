# Google Drive Block Device

```
go run *.go
mkfs.ext4 -O ^has_journal /dev/nbd0
mkdir /mnt/mem
mount /dev/nbd0 /mnt/mem
echo "Hello world" > /mnt/mem/hello.txt
cat /mnt/mem/README.md
ls /mnt/mem
```