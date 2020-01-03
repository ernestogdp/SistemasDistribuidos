package main

import (
	"fmt"
	"log"
	"sync"
	"time"
	"math/rand"
)

type bridgetype struct{
	crossCE int
	crossCW int
	waitCarE int
	waitCarW int
	finishCarE int
	finishCarW int
}

var (
	sem = make(chan struct{})
	limitsem = make(chan struct{})
	fin = make(chan struct{})
	mu sync.Mutex
	east sync.Mutex
	west sync.Mutex
	bridge bridgetype
	maxNumCars int = 20
	limitBridge int = 5
)

func randInt(min int, max int) int {
    rand.Seed(time.Now().UTC().UnixNano())
    return min + rand.Intn(max-min)
}

func sendFin(waitingcars int){
	if(waitingcars==0 && bridge.finishCarE+bridge.finishCarW == maxNumCars){
		fin <- struct{}{}
	}
}

func carsEast(id_carE int){

	var ihadwait = false
	i:=0

	mu.Lock()
	fmt.Printf("Coche [%d]: Llegada al puente en sentido  |Oeste <--- Este|\n", id_carE)
	//***** Comprobar si se puede pasar *****
	if(bridge.crossCW > 0){
		bridge.waitCarE++
		ihadwait = true
		mu.Unlock()
		<- sem
		mu.Lock()
	}

	//Cruzamos el puente
	if(ihadwait){
		bridge.waitCarE--
	}
	bridge.crossCE++
	mu.Unlock()

	//fmt.Printf("Coche [%d]: Cruzando el puente en sentido  |Oeste <--- Este|\n", id_carE)
	time.Sleep(1000*time.Millisecond)

	mu.Lock()
	bridge.finishCarE++
	bridge.crossCE--
	fmt.Printf("Coche [%d]: Ha cruzado el puente en sentido  |Oeste <--- Este|\n", id_carE)
	if(bridge.crossCE == 0 && bridge.waitCarE == 0 ){
		for i = 0; i<bridge.waitCarW; i++ {
			sem <- struct{}{}
		}
		//Solo se envia si es la última gorutina
		sendFin(i)
	}
	mu.Unlock()
}

func carsWest(id_carW int){

	var ihadwait = false
	i:=0

	mu.Lock()
	fmt.Printf("Coche [%d]: Llegada al puente en sentido  |Oeste ---> Este|\n", id_carW)

	//***** Comprobar si se puede pasar *****
	if(bridge.crossCE > 0){
		bridge.waitCarW++
		ihadwait = true
		mu.Unlock()
		<- sem
		mu.Lock()
	}

	//Cruzamos el puente
	if(ihadwait){
		bridge.waitCarW--
	}
	bridge.crossCW++
	mu.Unlock()

	//fmt.Printf("Coche [%d]: Cruzando el puente en sentido  |Oeste ---> Este|\n", id_carW)
	time.Sleep(1000*time.Millisecond)

	mu.Lock()
	bridge.finishCarW++
	bridge.crossCW--
	fmt.Printf("Coche [%d]: Ha cruzado el puente  en sentido  |Oeste ---> Este|\n", id_carW)

	if(bridge.crossCW == 0 && bridge.waitCarW == 0 ){
		for i=0; i<bridge.waitCarE; i++ {
			sem <- struct{}{}
		}
		sendFin(i)
	}
	mu.Unlock()

}

func createCars(){

	for car := 0; car < maxNumCars; car++ {
		choosedirection := randInt(0, 2)
		switch choosedirection {
		case 0:
			go carsWest(car)
			fmt.Printf("Coche [%d]: Salida desde el Oeste\n", car)
		case 1:
			go carsEast(car)
			fmt.Printf("Coche [%d]: Salida desde el Este\n", car)
		default:
			log.Print("Car loss\n")
		}
	}
}

//!-+main
func main() {

	go createCars()

	//Cuando recibimos la señal de que ha terminado la ultima gorutina imprimimos las variables compartidas
	<- fin
	fmt.Printf("\nFinalizados Oeste: %d  |  Esperando Oeste: %d   |  CruzandoW: %d\n", bridge.finishCarW, bridge.waitCarW, bridge.crossCW)
	fmt.Printf("Finalizados Este:  %d  |  Esperando Este:  %d   |  CruzandoE: %d\n\n", bridge.finishCarE, bridge.waitCarE, bridge.crossCE)

}
//!-main
