package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"syscall"
	"fmt"
)

func main() {
	outLog, err := execScripts(exec.Command("sh", "-c", "/sbin/fdisk -l 2> /dev/null | /usr/bin/awk '{print $1}' | /bin/grep /dev/.d "))
	if err != nil {
		log.Fatal(err)
	}
	var (
		a     syscall.Statfs_t
		max   uint64
		maxId string
	)
	outLogString := strings.Split(outLog, "\n")
	for _, e := range outLogString {
		_, err := execScripts(exec.Command("sh", "-c", "mount -r "+e+" /mnt"))
		if err == nil {
			syscall.Statfs("/mnt", &a)
			log.Println(a.Bfree * uint64(a.Bsize) / 1024)
			if a.Bfree*uint64(a.Bsize)/1024 > max {
				max = a.Bfree * uint64(a.Bsize) / 1024
				maxId = e
			}
			if err != nil {
				log.Fatal(err)
			}
			_, err = execScripts(exec.Command("sh", "-c", "sleep 1 && umount /mnt"))
			if err != nil {
				log.Println(err)
			}
		}
	}
	if float64(max)/1048576 >= 40.0 {
		log.Println("max size is :", float64(max)/1048576.0)
		fs, err := execScripts(exec.Command("sh", "-c", "lsblk -f "+maxId+" | awk 'NR==2{print $2}'"))
		if err != nil {
			log.Fatal(err)
		}
		if strings.TrimRight(string(fs), "\n") == "ntfs" {
			_, err = execScripts(exec.Command("sh", "-c", "size=$((($(sfdisk -s "+maxId+")-20971520)*1024)) && echo $size  && ntfsresize --size $size "+maxId+" && ntfsfix -d "+maxId))
			if err != nil {
				log.Fatal(err)
			}
		} else if strings.TrimRight(string(fs), "\n 1 2 3 4 5 6 7 8 9 0") == "ext" {
			execScripts(exec.Command("sh", "-c", "size=$((($(sfdisk -s "+maxId+")-20971520))) && echo $size  && resize2fs "+maxId+" $size\"k\" && e2fsck -f -y "+maxId))
		} else {
			fmt.Println("not supported fs")
		}
		r := strings.NewReplacer("/", "\\/")
		_, err = execScripts(exec.Command("sh", "-c", "sfdisk -d "+strings.TrimRight(maxId, "1234567890")+" > /opt/ptold.sfdisk && oldsize=$(($(sfdisk -s "+maxId+")*2)) && newsize=$(($oldsize-41943040)) && sed '"+r.Replace(maxId)+"/s/'$oldsize'/'$newsize'/g' /opt/ptold.sfdisk > /opt/ptnew.sfdisk && sfdisk "+strings.TrimRight(maxId, "1234567890")+" < /opt/ptnew.sfdisk"))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("not enough free space")
	}
}
func execScripts(cmd *exec.Cmd) (string, error) {
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}
	outLog, err := ioutil.ReadAll(stdout)
	if err != nil {
		return "", err
	}
	if len(outLog) != 0 {
		log.Println(string(outLog))
	}
	errLog, err := ioutil.ReadAll(stderr)
	if err != nil {
		return "", err
	}
	if len(errLog) != 0 {
		return "", errors.New(string(errLog))
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	return string(outLog), nil
}
