package gosocket

import (
	"encoding/json"
	"log"
)

func parseMessage(message Message) []byte {
	arrayByte, err := json.Marshal(message)
	if err != nil {
		log.Println("----------------Go Socket------------------")
		log.Println("Can not parse message to []byte")
		log.Println(message)
		log.Println(err)
		log.Println("----------------Go Socket------------------")
	}
	return arrayByte
}

func parseByteToMessage(arrayByte []byte) Message {
	message := Message{}
	err := json.Unmarshal(arrayByte, &message)
	if err != nil {
		log.Println("----------------Go Socket------------------")
		log.Println("Can not parse []byte to Message")
		log.Println(message)
		log.Println(err)
		log.Println("----------------Go Socket------------------")
	}
	return message
}
