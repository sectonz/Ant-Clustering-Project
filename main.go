package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const caminho_resultados = "resultados/"
const caminho_dataset = "dataset/4.txt"
const alpha = 0.5
const k1 = 0.1
const k2 = 0.15

type Dado struct {
	atributos []float64
	classe    int
}

type Formiga struct {
	isBusy        bool
	xAtual        int
	yAtual        int
	dadoCarregado *Dado
}

type ItemPlot struct {
	Row    int `json:"row"`
	Col    int `json:"col"`
	Classe int `json:"classe"`
}

type AntPlot struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type DadosPlot struct {
	GridHeight       int        `json:"grid_height"`
	GridWidth        int        `json:"grid_width"`
	NumAnts          int        `json:"num_ants"`
	NumItems         int        `json:"num_items"`
	Iterations       int        `json:"iterations"`
	VisionRadius     int        `json:"vision_radius"`
	K1               float64    `json:"k1"`
	K2               float64    `json:"k2"`
	Alpha            float64    `json:"alpha"`
	IterationCurrent int        `json:"iteration_current"`
	PhaseName        string     `json:"phase_name"`
	Items            []ItemPlot `json:"items"`
	Ants             []AntPlot  `json:"ants"`
}

func main() {
	var (
		mAltura    = 50
		mLargura   = 50
		numAgentes = 50
		raio       = 1
		iteracoes  = 2500000
	)

	if len(os.Args) >= 5 {
		mAltura = atoi(os.Args[1])
		mLargura = atoi(os.Args[2])
		numAgentes = atoi(os.Args[3])
		raio = atoi(os.Args[4])
		if len(os.Args) >= 6 {
			iteracoes = atoi(os.Args[5])
		}
	}

	simulate_ant_clustering(mAltura, mLargura, numAgentes, raio, iteracoes)
}

func atoi(s string) int {
	valor, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return valor
}

func load_dataset_normalizado(caminho string) []*Dado {
	f, err := os.Open(caminho)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var dados []*Dado

	for scanner.Scan() {
		linha := strings.TrimSpace(scanner.Text())
		if linha == "" || strings.HasPrefix(linha, "#") {
			continue
		}

		linha = strings.ReplaceAll(linha, ",", ".")
		campos := strings.Fields(linha)
		if len(campos) < 3 {
			continue
		}

		classe := atoi(campos[len(campos)-1])

		atributos := make([]float64, len(campos)-1)
		for i := 0; i < len(campos)-1; i++ {
			val, err := strconv.ParseFloat(campos[i], 64)
			if err != nil {
				log.Fatal(err)
			}
			atributos[i] = val
		}

		dados = append(dados, &Dado{atributos: atributos, classe: classe})
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	if len(dados) == 0 {
		log.Fatal("dataset vazio")
	}

	// normalizacao min-max para cada dimensao
	numAtributos := len(dados[0].atributos)
	minVals := make([]float64, numAtributos)
	maxVals := make([]float64, numAtributos)

	for j := 0; j < numAtributos; j++ {
		minVals[j] = dados[0].atributos[j]
		maxVals[j] = dados[0].atributos[j]
	}

	for _, d := range dados {
		for j := 0; j < numAtributos; j++ {
			if d.atributos[j] < minVals[j] {
				minVals[j] = d.atributos[j]
			}
			if d.atributos[j] > maxVals[j] {
				maxVals[j] = d.atributos[j]
			}
		}
	}

	for _, d := range dados {
		for j := 0; j < numAtributos; j++ {
			diff := maxVals[j] - minVals[j]
			if diff != 0 {
				d.atributos[j] = (d.atributos[j] - minVals[j]) / diff
			} else {
				d.atributos[j] = 0
			}
		}
	}

	return dados
}

func fill_matriz(
	mAltura int,
	mLargura int,
	dados []*Dado,
) [][]*Dado {

	matriz := make([][]*Dado, mAltura)
	for i := range matriz {
		matriz[i] = make([]*Dado, mLargura)
	}

	posicoes := rand.Perm(mAltura * mLargura)[:len(dados)]

	for idx, pos := range posicoes {
		x := pos / mLargura
		y := pos % mLargura
		matriz[x][y] = dados[idx]
	}

	return matriz
}

func create_swarm(
	numAgentes int,
	mAltura int,
	mLargura int,
	ocupada [][]bool,
) []Formiga {

	formigas := make([]Formiga, numAgentes)

	for i := range numAgentes {

		x := rand.Intn(mAltura)
		y := rand.Intn(mLargura)

		for ocupada[x][y] {
			x = rand.Intn(mAltura)
			y = rand.Intn(mLargura)
		}

		ocupada[x][y] = true

		formigas[i] = Formiga{isBusy: false, xAtual: x, yAtual: y, dadoCarregado: nil}

	}

	return formigas
}

func distancia_euclidiana(a, b []float64) float64 {
	soma := 0.0
	for k := 0; k < len(a); k++ {
		diff := a[k] - b[k]
		soma += diff * diff
	}
	return math.Sqrt(soma)
}

func calc_f(
	matriz [][]*Dado,
	dado *Dado,
	x int,
	y int,
	raio int,
) float64 {
	mAltura := len(matriz)
	mLargura := len(matriz[0])

	soma := 0.0
	s := 2*raio + 1
	s2 := float64(s * s)

	for dx := -raio; dx <= raio; dx++ {
		for dy := -raio; dy <= raio; dy++ {

			if dx == 0 && dy == 0 {
				continue
			}

			nx := (x + dx + mAltura) % mAltura
			ny := (y + dy + mLargura) % mLargura

			vizinho := matriz[nx][ny]
			if vizinho != nil {
				d := distancia_euclidiana(dado.atributos, vizinho.atributos)
				soma += 1.0 - (d / alpha)
			}
		}
	}

	f := soma / s2
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}

func prob_pegar(f float64) float64 {
	termo := k1 / (k1 + f)
	return termo * termo
}

func prob_largar(f float64) float64 {
	termo := f / (k2 + f)
	return termo * termo
}

// move cima/baixo/direita/esquerda, com movimento toroidal e checagem de colisao
func move_Formiga(
	x int,
	y int,
	mAltura int,
	mLargura int,
	ocupada [][]bool,
) (int, int) {

	direcoes := [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	perm := rand.Perm(4)

	for _, idx := range perm {
		d := direcoes[idx]
		nx := (x + d[0] + mAltura) % mAltura
		ny := (y + d[1] + mLargura) % mLargura

		if !ocupada[nx][ny] {
			ocupada[x][y] = false
			ocupada[nx][ny] = true
			return nx, ny
		}
	}

	// permanece no lugar se todas as posicoes vizinhas estiverem ocupadas
	return x, y
}

func save_matrix(
	caminho string,
	matriz [][]*Dado,
) {
	var sb strings.Builder

	for _, linha := range matriz {
		for j, dado := range linha {
			if j > 0 {
				sb.WriteString(" ")
			}
			valor := 0
			if dado != nil {
				valor = dado.classe
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

func gerar_grafico(
	matriz [][]*Dado,
	formigas []Formiga,
	mAltura int,
	mLargura int,
	numItems int,
	iteracoesTotal int,
	iteracaoAtual int,
	raio int,
	fase string,
	nomeBase string,
) {
	var items []ItemPlot
	for r := 0; r < mAltura; r++ {
		for c := 0; c < mLargura; c++ {
			if matriz[r][c] != nil {
				items = append(items, ItemPlot{
					Row:    r,
					Col:    c,
					Classe: matriz[r][c].classe,
				})
			}
		}
	}

	var ants []AntPlot
	for _, f := range formigas {
		ants = append(ants, AntPlot{
			Row: f.xAtual,
			Col: f.yAtual,
		})
	}

	plotData := DadosPlot{
		GridHeight:       mAltura,
		GridWidth:        mLargura,
		NumAnts:          len(formigas),
		NumItems:         numItems,
		Iterations:       iteracoesTotal,
		VisionRadius:     raio,
		K1:               k1,
		K2:               k2,
		Alpha:            alpha,
		IterationCurrent: iteracaoAtual,
		PhaseName:        fase,
		Items:            items,
		Ants:             ants,
	}

	jsonData, err := json.Marshal(plotData)
	if err != nil {
		log.Fatalf("erro ao serializar json do plot: %v", err)
	}

	jsonPath := caminho_resultados + nomeBase + ".json"
	pngPath := caminho_resultados + nomeBase + ".png"

	err = os.WriteFile(jsonPath, jsonData, 0644)
	if err != nil {
		log.Fatalf("erro ao salvar arquivo json: %v", err)
	}

	cmd := exec.Command("python3", "plot.py", jsonPath, pngPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("erro ao executar plot.py: %v, saida: %s", err, string(out))
	} else {
		os.Remove(jsonPath)
	}
}

func simulate_ant_clustering(
	mAltura int,
	mLargura int,
	numAgentes int,
	raio int,
	iteracoes int,
) {
	_ = os.MkdirAll(caminho_resultados, 0755)

	dados := load_dataset_normalizado(caminho_dataset)
	numItems := len(dados)

	matriz := fill_matriz(mAltura, mLargura, dados)

	ocupada := make([][]bool, mAltura)
	for i := range ocupada {
		ocupada[i] = make([]bool, mLargura)
	}

	formigas := create_swarm(numAgentes, mAltura, mLargura, ocupada)

	save_matrix(caminho_resultados+"inicio.txt", matriz)
	gerar_grafico(matriz, formigas, mAltura, mLargura, numItems, iteracoes, 0, raio, "Início", "inicio")

	for i := 0; i < iteracoes; i++ {

		for j := 0; j < numAgentes; j++ {

			x_atual := formigas[j].xAtual
			y_atual := formigas[j].yAtual

			if formigas[j].isBusy {
				// dropa ou nao dropa?
				if matriz[x_atual][y_atual] == nil {
					// nao tem item, pode dropar
					f := calc_f(matriz, formigas[j].dadoCarregado, x_atual, y_atual, raio)
					pd := prob_largar(f)

					if rand.Float64() < pd {
						matriz[x_atual][y_atual] = formigas[j].dadoCarregado
						formigas[j].dadoCarregado = nil
						formigas[j].isBusy = false
					}
				} else {
					// tem item, formiga nao empilha, continua carregando
				}

			} else {
				// pega ou nao pega?
				if matriz[x_atual][y_atual] == nil {
					// nao tem item, nada para pegar
				} else {
					// tem item, pode pegar
					dado := matriz[x_atual][y_atual]
					f := calc_f(matriz, dado, x_atual, y_atual, raio)
					pp := prob_pegar(f)

					if rand.Float64() < pp {
						formigas[j].dadoCarregado = dado
						matriz[x_atual][y_atual] = nil
						formigas[j].isBusy = true
					}
				}
			}

			formigas[j].xAtual, formigas[j].yAtual = move_Formiga(x_atual, y_atual, mAltura, mLargura, ocupada)

		}

		if i == iteracoes/2 {
			save_matrix(caminho_resultados+"meio.txt", matriz)
			gerar_grafico(matriz, formigas, mAltura, mLargura, numItems, iteracoes, i, raio, "Meio", "meio")
		}

	}

	save_matrix(caminho_resultados+"final.txt", matriz)
	gerar_grafico(matriz, formigas, mAltura, mLargura, numItems, iteracoes, iteracoes, raio, "Final", "final")

}
