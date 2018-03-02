package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	cmd := exec.Command("sh", "-c", "/sbin/fdisk -l 2> /dev/null | /usr/bin/awk '{print $1}' | /bin/grep /dev/.d ")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	outLog, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	outLogString := strings.Split(string(outLog), "\n")
	var max int
	var maxId string
	for _, e := range outLogString {
		mount(e)
		max, maxId,err=calcFreeSpace(e,max,maxId)
		if err != nil{
			log.Fatal(err)
		}
		umount(e)
	}
	if float64(max)/1048576 >= 40.0 {
		log.Println(float64(max) / 1048576.0)
	}

	err, fs := checkFs(maxId)
	shrinkVolume(maxId, fs)
	shrinkPartition(maxId)
}
func mount(id string) error {
	cmd := exec.Command("sh", "-c", "mount -r "+id+" /mnt"+"")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	errLog, err := ioutil.ReadAll(stderr)
	if err != nil {
		return err
	}
	if len(errLog) != 0 {
		errors.New(string(errLog))
	}
	return nil
}

func calcFreeSpace(id string, max int, maxId string) (int, string, error) {
	cmd := exec.Command("sh", "-c", "df "+id+" | awk 'NR==2{print $4}'")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 0, "", err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, "", err
	}
	if err := cmd.Start(); err != nil {
		return 0, "", err
	}
	outLog, err := ioutil.ReadAll(stdout)
	if err != nil {
		return 0, "", err
	}
	if len(outLog) != 0 {
		log.Println(string(outLog))
	}
	errLog, err := ioutil.ReadAll(stderr)
	if err != nil {
		return 0, "", err
	}
	if len(errLog) != 0 {
		return 0, "", errors.New(string(errLog))
	}
	i, err := strconv.Atoi(strings.TrimRight(string(outLog), "\n"))
	if err != nil {
		return 0, "", err
	}
	if max < i {
		return i, id, err
	}

	return max, maxId, err
}

func checkFs(id string) (error, string) {
	cmd := exec.Command("sh", "-c", "lsblk -f "+id+" | awk 'NR==2{print $2}'")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err, ""
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err, ""
	}

	if err := cmd.Start(); err != nil {
		return err, ""
	}
	outLog, err := ioutil.ReadAll(stdout)
	if err != nil {
		return err, ""
	}
	if len(outLog) != 0 {
		log.Println(string(outLog))
	}
	errLog, err := ioutil.ReadAll(stderr)
	if err != nil {
		return err, ""
	}
	if len(errLog) != 0 {
		return errors.New(string(errLog)), ""
	}
	return nil, strings.TrimRight(string(outLog), "\n")
}

func shrinkVolume(id string, fs string) error {
	cmd := exec.Command("sh", "-c", "echo \"not supported fs\" ")
	if fs == "ntfs" {
		cmd = exec.Command("sh", "-c", "size=$((($(sfdisk -s "+id+")-20971520)*1024)) && echo $size  && ntfsresize --no-action --size $size "+id+" && ntfsfix -d "+id+"&& ping -c4 127.0.0.1")
	} else if fs == "ext4" {
		cmd = exec.Command("sh", "-c", "echo \"ext support will become soon\" ")
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	outLog, err := ioutil.ReadAll(stdout)
	if err != nil {
		return err
	}
	if len(outLog) != 0 {
		log.Println(string(outLog))
	}
	errLog, err := ioutil.ReadAll(stderr)
	if err != nil {
		return err
	}
	if len(errLog) != 0 {
		return errors.New(string(errLog))
	}
	return nil
}

func shrinkPartition(id string) error {
	r := strings.NewReplacer("/", "\\/")
	cmd := exec.Command("sh", "-c", "sfdisk -d "+strings.TrimRight(id, "1234567890")+" > /opt/ptold.sfdisk && oldsize=$(($(sfdisk -s "+id+")*2)) && newsize=$(($oldsize-41943040)) && sed '"+r.Replace(id)+"/s/'$oldsize'/'$newsize'/g' /opt/ptold.sfdisk > /opt/ptnew.sfdisk && sfdisk -n "+strings.TrimRight(id, "1234567890")+" < /opt/ptnew.sfdisk")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	outLog, err := ioutil.ReadAll(stdout)
	if err != nil {
		return err
	}
	if len(outLog) != 0 {
		log.Println(string(outLog))
	}
	errLog, err := ioutil.ReadAll(stderr)
	if err != nil {
		return err
	}
	if len(errLog) != 0 {
		return errors.New(string(errLog))
	}
	return nil
}

func umount(id string) error {
	cmd := exec.Command("sh", "-c", "umount "+id)
	out, err := cmd.CombinedOutput()
	if err != nil { return err }
	log.Println(string(out))
	return nil
}
