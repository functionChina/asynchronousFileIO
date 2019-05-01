package asynchronousFileIO

import (
	"fmt"
	"os"
	"sync"
)

type unChangeDataManager struct {
	produceTable map[string]chan AFIOInterface
	produceLock *sync.RWMutex

	produceLine chan AFIOInterface
	saveRequest chan *saveCmd
	loadRequest chan *loadCmd
	model []AFIOInterface
	ds []DataSource
	logDistance *os.File
}

func (m *unChangeDataManager) saveWorker(data AFIOInterface, t uint64) {
	count := uint8(0)
retry:
	f, err := m.ds[t].WriteTo(data.GetFileName())
	if err != nil{
		_, err2 := fmt.Fprintln(m.logDistance, err.Error())
		if err != nil{
			fmt.Println(err2.Error())
			os.Exit(20010) // log fail
		}
		if count < 3{
			goto retry
		}
	}
	err = data.Save(f)
	if err != nil{
		_, err2 := fmt.Fprintln(m.logDistance, err.Error())
		if err != nil{
			fmt.Println(err2.Error())
			os.Exit(20010) // log fail
		}
		if count < 3{
			goto retry
		}
	}
	err = m.ds[t].CloseWrite(f)
	if err != nil{
		_, err2 := fmt.Fprintln(m.logDistance, err.Error())
		if err != nil{
			fmt.Println(err2.Error())
			os.Exit(20010) // log fail
		}
		if count < 3{
			goto retry
		}
	}
}

func (m *unChangeDataManager) loadWorker(fileName string, t uint64) {
	var count uint8
retry:
	f, err := m.ds[t].ReadFrom(fileName)
	if err != nil{
		if err != nil{
			_, err2 := fmt.Fprintln(m.logDistance, err.Error())
			if err != nil{
				fmt.Println(err2.Error())
				os.Exit(20010) // log fail
			}
			if count < 3{
				goto retry
			}
		}
	}
	goal := m.model[t].Clone()
	err = goal.Load(f)
	if err != nil{
		if err != nil{
			_, err2 := fmt.Fprintln(m.logDistance, err.Error())
			if err != nil{
				fmt.Println(err2.Error())
				os.Exit(20010) // log fail
			}
			if count < 3{
				goto retry
			}
		}
	}
	err = m.ds[t].CloseReader(f)
	if err != nil{
		if err != nil{
			_, err2 := fmt.Fprintln(m.logDistance, err.Error())
			if err != nil{
				fmt.Println(err2.Error())
				os.Exit(20010) // log fail
			}
			if count < 3{
				goto retry
			}
		}
	}
	m.produceLine <- goal
}

func (m *unChangeDataManager) Load(fileName string, t uint64){
	m.loadRequest <- &loadCmd{t, fileName}
}

func (m *unChangeDataManager) Save(data AFIOInterface, t uint64) {
	m.saveRequest <- &saveCmd{t, data}
}

func (m *unChangeDataManager) GetData(fileName string) (goal AFIOInterface) {
	m.produceLock.RLock()
	if c, exist := m.produceTable[fileName]; !exist{
		m.produceLock.RUnlock()
		return nil
	} else {
		m.produceLock.RUnlock()
		goal = <- c
		m.produceLock.Lock()
		delete(m.produceTable, fileName)
		m.produceLock.Unlock()
		return
	}
}

func (m *unChangeDataManager) fileManager(){
	for{
		select {
		case p:=<-m.produceLine:
			m.produceLock.RLock()
			m.produceTable[p.GetFileName()] <- p
			m.produceLock.RUnlock()
		case p:=<-m.saveRequest:
			go m.saveWorker(p.data, p.t)
		case c := <-m.loadRequest:
			m.produceLock.RLock()
			if _, ok := m.produceTable[c.fileName]; ok{
				m.produceLock.RUnlock()
				break
			}
			m.produceLock.Lock()
			m.produceTable[c.fileName] = make(chan AFIOInterface)
			m.produceLock.Unlock()
			go m.loadWorker(c.fileName, c.t)
		}
	}
}

// if you don't want to had a special log information distance you can pass nil to the third parameter it will auto to console
func NewUnChangeFileManager(model []AFIOInterface, source []DataSource, logDistance *os.File) (goal FileManager){
	if logDistance == nil{
		logDistance = os.Stdout
	}
	goal = &unChangeDataManager{make(map[string]chan AFIOInterface),
		new(sync.RWMutex),
		make(chan AFIOInterface, 20),
		make(chan *saveCmd, 20),
		make(chan *loadCmd, 20),model,
		source,
		logDistance}
	go goal.(*unChangeDataManager).fileManager()
	return
}

