package commiter

import (
	"encoding/binary"
	"errors"
	"global"
	"io"
	"message"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
)

/*
 存储格式：
 1.dat, 2.dat, ..., n.dat 存储commit的记录，每个文件存放400w条(0xFFFFFFFF)
 1.off, 2.off, ..., n.off 存储对应的record在dat文件中的off uint64
 1.con, 2.con, ..., n.con 消费记录
*/

type DiskCommiter struct {
	Commiter

	path  string
	curId uint64
	maxId uint64

	commitLock  sync.Mutex
	consumeLock sync.Mutex

	wFileId  uint64
	wDatFile *os.File
	wOffFile *os.File
	wConFile *os.File

	rFileId  uint64
	rDatFile *os.File

	mask uint64
}

/*
新建一个DiskCommiter，
path必须不存在
如果path存在，应该使用RecoverDiskCommiter进行恢复
*/
func NewDiskCommiter(path string) *DiskCommiter {
	err := os.Mkdir(path, 0777)
	if err != nil {
		panic(err)
	}
	if path[len(path)-1] != '/' {
		path += "/"
	}
	d := new(DiskCommiter)
	d.path = path
	d.curId = 0
	d.maxId = 0

	d.wFileId = 0
	rFlag := os.O_RDONLY
	wFlag := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	d.wDatFile, err = os.OpenFile(path+"0.dat", wFlag, 0666)
	if err != nil {
		goto error1
	}
	d.wOffFile, err = os.OpenFile(path+"0.off", wFlag, 0666)
	if err != nil {
		goto error2
	}
	d.wConFile, err = os.OpenFile(path+"0.con", wFlag, 0666)
	if err != nil {
		goto error3
	}

	d.rFileId = 0
	d.rDatFile, err = os.OpenFile(path+"0.dat", rFlag, 0666)
	if err != nil {
		goto error4
	}
	d.mask = 0xFFFFFFFF
	return d

error4:
	os.Remove(path + "0.con")
error3:
	os.Remove(path + "0.off")
error2:
	os.Remove(path + "0.dat")
error1:
	os.RemoveAll(path)
	panic(err)
}

func RecoverDiskCommiter(dataPath string) *DiskCommiter {
	if dataPath[len(dataPath)-1] != '/' {
		dataPath += "/"
	}
	max := -1
	var conFile string
	var conId int
	filepath.Walk(dataPath, func(path_ string, info os.FileInfo, err error) error {
		var num, idx int
		if err != nil {
			panic(err)
		}
		/* 获取扩展名 */
		ext := path.Ext(path_)
		var name string
		switch ext {
		case ".off":
			_, name = path.Split(path_)
			idx = len(name) - 4 /* xxx.off */
			num, err = strconv.Atoi(name[:idx])
			if err != nil {
				panic(err)
			}
			if num > max {
				max = num
			}
		case ".con":
			conFile = path_
			_, name = path.Split(path_)
			idx = len(name) - 4 /* xxx.con */
			conId, err = strconv.Atoi(name[:idx])
			if err != nil {
				panic(err)
			}
		}
		return nil
	})
	maxOffFile := dataPath + strconv.Itoa(max) + ".off"
	maxDatFile := dataPath + strconv.Itoa(max) + ".dat"

	d := new(DiskCommiter)
	d.path = dataPath
	d.mask = 0xFFFFFFFF
	d.wFileId = uint64(max)
	d.rFileId = uint64(conId)
	wFlag := os.O_APPEND | os.O_WRONLY
	rFlag := os.O_RDONLY

	var off int64
	var err error

	/* Append DataFile */
	d.wDatFile, err = os.OpenFile(maxDatFile, wFlag, 0666)
	if err != nil {
		panic(err)
	}

	/* Append OffFile, calc max msg Id */
	d.wOffFile, err = os.OpenFile(maxOffFile, wFlag, 0666)
	if err != nil {
		panic(err)
	}
	off, err = d.wOffFile.Seek(0, os.SEEK_END)
	if err != nil {
		panic(err)
	}
	d.maxId = ((d.mask + 1) * uint64(max)) + (uint64(off) >> 3)

	/* Append ConFile, calc current msg Id */
	d.wConFile, err = os.OpenFile(conFile, wFlag, 0666)
	if err != nil {
		panic(err)
	}
	off, err = d.wConFile.Seek(0, os.SEEK_END)
	if err != nil {
		panic(err)
	}
	d.curId = ((d.mask + 1) * uint64(conId)) + (uint64(off) >> 3)

	/* Open rDatFile and Seek to cur read pos */
	rDatFileName := dataPath + strconv.Itoa(conId) + ".dat"
	rOffFileName := dataPath + strconv.Itoa(conId) + ".off"

	d.rDatFile, err = os.OpenFile(rDatFileName, rFlag, 0666)
	if err != nil {
		panic(err)
	}

	var rOffFile *os.File
	rOffFile, err = os.OpenFile(rOffFileName, rFlag, 0666)
	if err != nil {
		panic(err)
	}
	defer rOffFile.Close()

	_, err = rOffFile.Seek(8*int64(d.curId&d.mask), os.SEEK_SET)
	if err != nil {
		panic(err)
	}
	err = binary.Read(rOffFile, binary.LittleEndian, &off)
	if err != nil {
		if d.curId == d.maxId && err == io.EOF {
			// 当前所有消息已被消费完
			_, err = d.rDatFile.Seek(0, os.SEEK_END)
		}
		if err != nil {
			panic(err)
		}
	} else {
		_, err = d.rDatFile.Seek(off, os.SEEK_SET)
		if err != nil {
			panic(err)
		}
	}

	return d
}

func (d *DiskCommiter) Total() uint64 {
	return d.maxId - d.curId
}

func (d *DiskCommiter) CurId() uint64 {
	return d.curId
}

func (d *DiskCommiter) MaxId() uint64 {
	return d.maxId
}

type MsgHdr struct {
	Id   uint64
	Size uint16
}

func (d *DiskCommiter) flush(msg message.Msg) (err error) {
	str, ok := msg.Content.(string)
	if !ok {
		return errors.New("Unkown msg.Content type")
	}

	/* write message off */
	var off int64
	off, err = d.wDatFile.Seek(0, os.SEEK_CUR)
	if err != nil {
		return
	}
	err = binary.Write(d.wOffFile, binary.LittleEndian, uint64(off))
	if err != nil {
		return
	}

	/* write message hdr */
	var hdr MsgHdr
	hdr.Id = msg.MsgId
	hdr.Size = uint16(len(str))
	err = binary.Write(d.wDatFile, binary.LittleEndian, hdr)
	if err != nil {
		global.Log("write off ok but write data file error, msgId:" + strconv.Itoa(int(msg.MsgId)))
		return
	}

	/* write message body */
	err = binary.Write(d.wDatFile, binary.LittleEndian, []byte(str))
	if err != nil {
		global.Log("write off and head ok but write body error, msgId:" + strconv.Itoa(int(msg.MsgId)))
		return
	}
	return
}

func (d *DiskCommiter) newRdDataFile() {
	d.rDatFile.Close()
	d.wConFile.Close()
	d.rFileId++
	rFlag := os.O_RDONLY
	wFlag := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	path := d.path + strconv.FormatUint(d.rFileId, 10) + ".dat"
	var err error
	d.rDatFile, err = os.OpenFile(path, rFlag, 0666)
	if err != nil {
		panic(err)
	}
	path = d.path + strconv.FormatUint(d.rFileId, 10) + ".con"
	d.wConFile, err = os.OpenFile(path, wFlag, 0666)
	if err != nil {
		panic(err)
	}
	os.Remove(d.path + strconv.FormatUint(d.rFileId-1, 10) + ".dat")
	os.Remove(d.path + strconv.FormatUint(d.rFileId-1, 10) + ".off")
	os.Remove(d.path + strconv.FormatUint(d.rFileId-1, 10) + ".con")
}

func (d *DiskCommiter) newWrDataFile() {
	d.wDatFile.Close()
	d.wOffFile.Close()
	d.wFileId++
	wFlag := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	path := d.path + strconv.FormatUint(d.wFileId, 10) + ".dat"
	var err error
	d.wDatFile, err = os.OpenFile(path, wFlag, 0666)
	if err != nil {
		panic(err)
	}
	path = d.path + strconv.FormatUint(d.wFileId, 10) + ".off"
	d.wOffFile, err = os.OpenFile(path, wFlag, 0666)
	if err != nil {
		panic(err)
	}
}

func (d *DiskCommiter) Commit(content interface{}) error {
	d.commitLock.Lock()
	defer d.commitLock.Unlock()

	var msg message.Msg
	msg.MsgId = d.maxId
	msg.Content = content
	err := d.flush(msg)
	if err != nil {
		return err
	}
	d.maxId++
	if d.maxId&d.mask == 0 {
		d.newWrDataFile()
	}
	return nil
}

func (d *DiskCommiter) GetMsg(id uint64) (msg interface{}, err error) {
	// TODO
	msg = nil
	err = nil
	return
}

func (d *DiskCommiter) Consume() (msg interface{}, err error) {
	if d.curId == d.maxId {
		err = errors.New("No more message")
		return
	}

	d.consumeLock.Lock()
	defer d.consumeLock.Unlock()

	var hdr MsgHdr
	var realMsg message.Msg

	/* Read Msg Header */
	err = binary.Read(d.rDatFile, binary.LittleEndian, &hdr)
	if err != nil {
		return
	}

	/* Read Msg Body */
	body := make([]byte, hdr.Size)
	err = binary.Read(d.rDatFile, binary.LittleEndian, body)
	if err != nil {
		return
	}

	/* Write con File */
	err = binary.Write(d.wConFile, binary.LittleEndian, d.curId)
	if err != nil {
		return
	}

	d.curId++
	realMsg.MsgId = hdr.Id
	realMsg.Content = string(body)
	msg = realMsg

	if d.curId&d.mask == 0 {
		d.newRdDataFile()
	}

	return
}
