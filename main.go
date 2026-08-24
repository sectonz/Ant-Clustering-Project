package main

import (
	"math/rand"
)

const caminho_resultados = "/resultados/"

type Formiga struct {
	isBusy bool
	xAtual int
	yAtual int
}

func main() {

	var (
		mAltura    = 50
		mLargura   = 50
		numAgentes = 15
		raio       = 1
		qtdItens   = 600
		iteracoes  = 100000
	)

	simulate_ant_clustering(mAltura, mLargura, numAgentes, raio, qtdItens, iteracoes)

}

func fill_matriz(mAltura int, mLargura int, qtdItens int) [][]int {

	matriz := make([][]int, mAltura)
	var total_itens int = 0

	for i := 0; i < mAltura; i++ {
		matriz[i] = make([]int, mLargura)
		for j := 0; j < mLargura; j++ {

			if total_itens != qtdItens {
				matriz[i][j] = rand.Intn(2)

				if matriz[i][j] == 1 {
					total_itens++
				}

			} else {
				matriz[i][j] = 0
			}
		}
	}

	return matriz
}

func create_swarm(matriz [][]int, numAgentes int,
	mAltura int, mLargura int) []Formiga {

	formigas := make([]Formiga, numAgentes)

	for i := range numAgentes {

		x := rand.Intn(mAltura)
		y := rand.Intn(mLargura)

		for matriz[x][y] == 8 {
			x = rand.Intn(mAltura)
			y = rand.Intn(mLargura)
		}

		matriz[x][y] = 8

		formigas[i] = Formiga{isBusy: false, xAtual: x, yAtual: y}

	}

	return formigas
}

func simulate_ant_clustering(mAltura int, mLargura int, numAgentes int,
	raio int, qtdItens int, iteracoes int) {

	matriz := fill_matriz(mAltura, mLargura, qtdItens)

	formigas := create_swarm(matriz, numAgentes, mAltura, mLargura)

	for i := 0; i < iteracoes; i++ {

		for j := 0; j < numAgentes; j++ {

			x_atual := formigas[j].xAtual
			y_atual := formigas[j].yAtual

			//desse jeito tem que começar andando, se não nunca vai ser 0 na celula atual
			//já que preencho com 8 (formiga)
			if formigas[j].isBusy {
				//dropa ou nao dropa?
				if matriz[x_atual][y_atual] == 0 {
					//n tem item
				} else {
					// tem item
				}

			} else {
				//pega ou nao pega?
				if matriz[x_atual][y_atual] == 0 {
					// nao tem item
				} else {
					// tem item
				}
			}

		}

	}

}
