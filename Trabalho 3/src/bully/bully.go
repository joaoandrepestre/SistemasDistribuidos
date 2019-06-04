package main

import (
	"fmt"
	"time"
	"strings"
)

var id int
var electionId int
var leaderId int

func main() {

	c := make(chan string)
	
	go receiveMsg(c)

	for i:=1;i<6;i++{
		c <- fmt.Sprintf("%d|3|19|000", i)
	}
	close(c)
	
}

func userInterface(){
	for{
		fmt.Printf("Interface com o usuário\n")
		time.Sleep(time.Second * 1)
	}
}

func receiveMsg(c chan string){
	var msg string
	var splitMsg []string
	for{
		msg = <- c
		splitMsg = strings.Split(msg, "|")
		msgType := splitMsg[0]


		switch msgType {
		case "1":
			fmt.Println("Eleição")
		case "2":
			fmt.Println("OK")
		case "3":
			fmt.Println("Líder")
		case "4":
			fmt.Println("Vivo")
		case "5":
			fmt.Println("Vivo_OK")
		}
	}
}

func detectLeader(){
	for{
		fmt.Printf("Exporadicamente checa se o líder está ativo\n")
		time.Sleep(time.Second * 1)
	}
}
