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

func fill_matriz(mAltura int, mLargura int) [][]int {

	matriz := make([][]int, mAltura)

	for i := 0; i < mAltura; i++ {
		matriz[i] = make([]int, mLargura)
		for j := 0; j < mLargura; j++ {
			matriz[i][j] = rand.Intn(2)
		}
	}

	return matriz
}

func create_swarm(matriz [][]int, numAgentes int,
	mAltura int, mLargura int) []Formiga {

	formigas := make([]Formiga, numAgentes)

	for i := 0; i < numAgentes; i++ {

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

	matriz := fill_matriz(mAltura, mLargura)

	formigas := create_swarm(matriz, numAgentes, mAltura, mLargura)

	for i := 0; i < iteracoes; i++ {

	}

}
