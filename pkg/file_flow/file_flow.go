package fileflow

import (
	"log"
	"os"
)

func (tf *TaskFile) read() {
	file, err := os.Open(tf.path + "/" + tf.name)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
}
