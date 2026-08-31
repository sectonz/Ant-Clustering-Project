package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

const caminho_resultados = "resultados/"

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

	if len(os.Args) >= 2 {
		mAltura = atoi(os.Args[1])
		mLargura = atoi(os.Args[2])
		numAgentes = atoi(os.Args[3])
		raio = atoi(os.Args[4])
		qtdItens = atoi(os.Args[5])
		iteracoes = atoi(os.Args[6])
	}

	simulate_ant_clustering(mAltura, mLargura, numAgentes, raio, qtdItens, iteracoes)

}

func atoi(s string) int {
	valor, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return valor
}

func fill_matriz(
	mAltura int,
	mLargura int,
	qtdItens int,
) [][]int {

	matriz := make([][]int, mAltura)
	for i := range matriz {
		matriz[i] = make([]int, mLargura)
	}

	// para deixar o preenchimento mais homogeneo e alatorio,
	// faz a permutacao de todas as posicoes e pega as primeiras qtdItens
	posicoes := rand.Perm(mAltura * mLargura)[:qtdItens]

	for _, pos := range posicoes {
		x := pos / mLargura
		y := pos % mLargura
		matriz[x][y] = 1
	}

	return matriz
}

func create_swarm(
	numAgentes int,
	mAltura int,
	mLargura int,
) []Formiga {

	formigas := make([]Formiga, numAgentes)

	ocupada := make([][]bool, mAltura)
	for i := range ocupada {
		ocupada[i] = make([]bool, mLargura)
	}

	for i := range numAgentes {

		x := rand.Intn(mAltura)
		y := rand.Intn(mLargura)

		for ocupada[x][y] {
			x = rand.Intn(mAltura)
			y = rand.Intn(mLargura)
		}

		ocupada[x][y] = true

		formigas[i] = Formiga{isBusy: false, xAtual: x, yAtual: y}

	}

	return formigas
}

func count_neighbors(
	matriz [][]int,
	x int,
	y int,
	raio int,
) (q int, c int) {

	mAltura := len(matriz)
	mLargura := len(matriz[0])

	for dx := -raio; dx <= raio; dx++ {
		for dy := -raio; dy <= raio; dy++ {

			if dx == 0 && dy == 0 {
				continue
			}

			nx := x + dx
			ny := y + dy

			// vizinho que passa do limite reaparece do outro lado
			nx = (nx + mAltura) % mAltura
			ny = (ny + mLargura) % mLargura

			c++
			if matriz[nx][ny] == 1 {
				q++
			}
		}
	}

	return q, c
}

// move cima/baixo/direita/esquerda, com movimento toroidal
func move_Formiga(
	x int,
	y int,
	mAltura int,
	mLargura int,
) (int, int) {

	direcoes := [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	d := direcoes[rand.Intn(4)]

	nx := x + d[0]
	ny := y + d[1]

	nx = (nx + mAltura) % mAltura
	ny = (ny + mLargura) % mLargura

	return nx, ny
}

func save_matrix(
	caminho string,
	matriz [][]int,
) {
	var sb strings.Builder

	for _, linha := range matriz {
		for j, valor := range linha {
			// espaco entre os elementos da linha, mas só apartir do segundo elemento da linha
			if j > 0 {
				sb.WriteString(" ")
			}
			fmt.Fprintf(&sb, "%d", valor)
		}
		sb.WriteString("\n")
	}

	err := os.WriteFile(caminho, []byte(sb.String()), 0644)
	if err != nil {
		panic(err)
	}
}

func register_on_the_fly(
	caminho string,
	matriz [][]int,
	iteracao int,
) {
	arquivo, err := os.OpenFile(caminho, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer arquivo.Close()

	fmt.Fprintf(arquivo, "=== iteracao %d ===\n", iteracao)

	for _, linha := range matriz {
		for j, valor := range linha {
			if j > 0 {
				fmt.Fprint(arquivo, " ")
			}
			fmt.Fprintf(arquivo, "%d", valor)
		}
		fmt.Fprintln(arquivo)
	}
}

func simulate_ant_clustering(
	mAltura int,
	mLargura int,
	numAgentes int,
	raio int,
	qtdItens int,
	iteracoes int,
) {
	matriz := fill_matriz(mAltura, mLargura, qtdItens)

	formigas := create_swarm(numAgentes, mAltura, mLargura)

	save_matrix(caminho_resultados+"inicio.txt", matriz)

	const intervaloOnTheFly = 1000

	for i := 0; i < iteracoes; i++ {

		for j := 0; j < numAgentes; j++ {

			x_atual := formigas[j].xAtual
			y_atual := formigas[j].yAtual

			if formigas[j].isBusy {
				//dropa ou nao dropa?
				if matriz[x_atual][y_atual] == 0 {
					// nao tem item, pode dropar
					q, c := count_neighbors(matriz, x_atual, y_atual, raio)
					fd := float64(q) / float64(c)

					if rand.Float64() < fd {
						matriz[x_atual][y_atual] = 1
						formigas[j].isBusy = false
					}
				} else {
					// tem item, formiga nao empilha, continua carregando
				}

			} else {
				//pega ou nao pega?
				if matriz[x_atual][y_atual] == 0 {
					// nao tem item, nada para pegar
				} else {
					// tem item, pode pegar
					q, c := count_neighbors(matriz, x_atual, y_atual, raio)
					fp := 1 - float64(q)/float64(c)

					if rand.Float64() < fp {
						matriz[x_atual][y_atual] = 0
						formigas[j].isBusy = true
					}
				}
			}

			formigas[j].xAtual, formigas[j].yAtual = move_Formiga(x_atual, y_atual, mAltura, mLargura)

		}

		if i == iteracoes/2 {
			save_matrix(caminho_resultados+"meio.txt", matriz)
		}

		if i%intervaloOnTheFly == 0 {
			register_on_the_fly(caminho_resultados+"on-the-fly.txt", matriz, i)
		}

	}

	save_matrix(caminho_resultados+"final.txt", matriz)

}
