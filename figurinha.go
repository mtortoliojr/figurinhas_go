package main

import (
	"fmt"
	"math/rand"
	"sync"
)

type Pacote map[int]int
type Pacotes []Pacote
type Exemplar map[int]int
type Figura map[int]int
type Figuras []int

// type Album map[int]int

type Album struct {
	Id        int
	Pacotes   int
	Tamanho   int
	Completo  int
	Figuras   Figura
	Repetidas []int
}

type Cliente struct {
	Id          int
	Completado  bool
	Recebimento chan Resposta
}

type Resposta struct {
	Esgotado bool
	Pacote   Pacote
}

func gerar_pacotes(tamanho_pacote int, exemplares Exemplar) Pacotes {

	pacotes := Pacotes{}
	album_tamanho := len(exemplares)
	n_exemplares := exemplares[0]

	figuras_disponiveis := album_tamanho * n_exemplares
	slots_disponiveis := album_tamanho

	for slots_disponiveis >= tamanho_pacote {
		pacote := make(Pacote)
		n_figurinhas := 0
		for n_figurinhas < tamanho_pacote {
			fig := int(rand.Int31n(int32(album_tamanho)) + 1)
			if exemplares[fig] > 0 {
				_, ok := pacote[fig]
				if !ok {
					pacote[fig] = 1
					n_figurinhas++
					exemplares[fig]--
					figuras_disponiveis--
				}
			}
		}

		slots_disponiveis = 0
		for _, n := range exemplares {
			if n > 0 {
				slots_disponiveis++
			}
		}

		pacotes = append(pacotes, pacote)
	}

	return pacotes
}

func gerar_exemplares(tamanho_album, numero_exemplares int) Exemplar {
	var exemplares = make(Exemplar)
	for i := 0; i < tamanho_album; i++ {
		exemplares[i+1] = numero_exemplares
	}

	return exemplares
}

func atualiza_album(album *Album, pacote Pacote) {
	for fig := range pacote {
		_, ok := album.Figuras[fig]
		if !ok {
			album.Completo++
		}
		album.Figuras[fig]++
	}
}

func comprar_album(tamanho_album int) Album {

	Figuras := make(Figura)

	albumNovo := Album{
		Tamanho: tamanho_album,
		Figuras: Figuras,
	}
	return albumNovo
}

func completar_album(id, tamanho_album int, banca chan<- Cliente, wg *sync.WaitGroup) {
	defer wg.Done()

	album := comprar_album(tamanho_album)

	cliente := Cliente{
		Id:          id,
		Completado:  false,
		Recebimento: make(chan Resposta),
	}

	for !(album.Completo == album.Tamanho) {

		// Se apresentar à banca
		banca <- cliente

		// Esperar o recebimento do pacote
		resposta := <-cliente.Recebimento

		if resposta.Esgotado {
			fmt.Printf("Acabaram os pacotes para o cliente %d :(: %d\n", cliente.Id, album.Pacotes)
			return
		}

		atualiza_album(&album, resposta.Pacote)
		album.Pacotes++
	}

	fmt.Printf("Cliente %d completou o album :): %d\n", cliente.Id, album.Pacotes)
}

func vender_pacotes(pacotes Pacotes, banca <-chan Cliente) {
	for {
		ordem := <-banca
		if len(pacotes) == 0 {
			ordem.Recebimento <- Resposta{Esgotado: true}
		} else if ordem.Completado {
			fmt.Println("Fechando a banca")
			ordem.Recebimento <- Resposta{}
			return
		} else {
			pacote := pacotes[0]
			pacotes = pacotes[1:]
			resposta := Resposta{
				Esgotado: false,
				Pacote:   pacote,
			}
			ordem.Recebimento <- resposta
		}
	}
}

func main() {

	n_pessoas := 10
	var tamanho_album int = 100
	var numero_exemplares int = 40
	var pacote_tamanho int = 4

	exemplares := gerar_exemplares(tamanho_album, numero_exemplares)
	pacotes := gerar_pacotes(pacote_tamanho, exemplares)

	banca := make(chan Cliente)

	var wg sync.WaitGroup
	wg.Add(n_pessoas)

	go vender_pacotes(pacotes, banca)

	for i := 0; i < n_pessoas; i++ {
		go completar_album(i+1, tamanho_album, banca, &wg)
	}

	wg.Wait()

	cliente := Cliente{Completado: true, Recebimento: make(chan Resposta)}
	banca <- cliente

	<-cliente.Recebimento
	fmt.Println("Banca fechada")
}
