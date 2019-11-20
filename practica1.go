// Chat is a server that lets clients chat with each other.
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

//!+broadcaster
type client struct {
	canal chan<- string
	id string
	conection net.Conn
}

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string) // all incoming client messages
	clients = make(map[client]bool) // all connected clients
)

//Mejora 3: Envio Multicast
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

//Mejora 1: Informacion lista de clientes
func writeClientsList (listclient map[client]bool, ch chan string){
	i:=1

	ch <- "\nLista de usuarios conectados:"
	ch <- "-----------------------------"
	for c:= range listclient{
		ch <- strconv.Itoa(i) +"- " + c.id
		i++
	}
	ch <- "-----------------------------\n"
}

func broadcaster() {
	for {
		select {
		case msg := <-messages:
			// Broadcast incoming message to all
			// clients' outgoing message channels.
			for cli := range clients {
				//Mejora 3: Envio multicast
				if(!hasPrefix(msg, cli.id)){
					cli.canal <- msg
				}
			}

		case cli := <-entering:
			clients[cli] = true
			go writeClientsList(clients, messages)

		case cli := <-leaving:
			delete(clients, cli)
			close(cli.canal)
			go writeClientsList(clients, messages)
		}
	}
}

//!-broadcaster
//Mejora 2: Nombres de clientes legibles
func setUniqueId(ch chan string, conn net.Conn) string {

	var who string
	for{
		retry := false
		ch <- "\nPor favor, introduce un Nickname:"
		scanner := bufio.NewScanner(conn)
		scanner.Scan()
		who= scanner.Text()
		for cli := range clients{
			if(cli.id == who){
				ch <- "El nick [" + who + "] ya está ocupado, por favor elija otro..."
				retry = true
				break
			}
		}
		if(!retry){
			break
		}
	}
	ch <- "\nTu nickname es " + who
	messages <- "\n" + who + " se ha conectado"
	return who

}

func setPrivateAdress(id string, canal chan string, conn net.Conn) net.Conn{

	scanner := bufio.NewScanner(conn)

	for{
		canal <- "\n¿A quién quiere enviar mensajes por privado? Escriba un nickname."
		//Escribimos la lista de usuario solo al cliente
		writeClientsList(clients, canal)
		scanner.Scan()
		address:= scanner.Text()
		for c:= range clients{
			if (c.id==address){
				canal <- "\n===== Envie mensajes a " + address + " por privado ====="
				return c.conection
			}
		}
		canal <- "Error: El destinatario que ha introducido no existe. Vuelva a intentarlo..."
	}
}

func setMode(ch chan string, conn net.Conn, who string){
	scanner := bufio.NewScanner(conn)
	var exitmode bool

	for {
		exitmode=false
		ch <- "\nElija un modo para el chat: [1, 2]"
		ch <- "\t1: Envie mensajes a todos los clientes activos."
		ch <- "\t2: Envie mensajes a un cliente por privado."
		scanner.Scan()
		i, _ := strconv.Atoi(scanner.Text())
		switch i {
		case 1:
			ch <- "\n===== Envie mensajes a todos los clientes ====="
			ch <- "#NOTA: Puede salir al menu de seleccion de modo escribiendo 'exit'"
			for scanner.Scan() {
				if(scanner.Text()=="exit"){
					exitmode=true
					break
				}
				messages <- who + ": " + scanner.Text()
			}
			if(exitmode){
				continue
			}

		case 2:
			addressconn:=setPrivateAdress(who, ch, conn)
			chprivate:= make(chan string)
			go clientWriter(addressconn, chprivate)
			ch <- "#NOTA: Puede salir al menu de seleccion de modo escribiendo 'exit'"
			for scanner.Scan() {
				if(scanner.Text()=="exit"){
					exitmode=true
					break
				}
				chprivate <- "#[Mensaje Privado] " + who + ": " + scanner.Text()
			}
			if(exitmode){
				continue
			}

		default:
			ch <- "El valor introducido no es correcto."
			ch <- "Introduzca [1] o [2]"
			continue
		}
		break
	}
}

//!+handleConn
func handleConn(conn net.Conn) {
	ch := make(chan string) // outgoing client messages
	go clientWriter(conn, ch)

	//Mejora 2: Nombres de clientes legibles
	//Solo se aceptan identificadores que no estén escogidos
	who:=setUniqueId(ch, conn)

	entering <- client{ch, who, conn}
	time.Sleep(100 * time.Millisecond)

	//Mejora 4: Canales privados
	setMode(ch, conn, who)

	messages <- "\n" + who + " se ha desconectado"
	leaving <- client{ch, who, conn}
	conn.Close()

}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg) // NOTE: ignoring network errors
	}
}

//!-handleConn

//!+main
func main() {
	listener, err := net.Listen("tcp", "10.1.147.225:8000")
	if err != nil {
		log.Fatal(err)
	}

	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

//!-main
