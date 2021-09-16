package main

import (
	Deep "github.com/patrikeh/go-deep"
	Training "github.com/patrikeh/go-deep/training"
	Log "log"
)

// https://github.com/patrikeh/go-deep

//
// на выходе - изменение цены в % следующую секунду...
/*
но если на вход подавать лишь текущий стакан - ведь это предсказание следующей точки, не более - норм?

иначе говоря:
 - мы берем кусочек картинки за 1 минуту (=100ms, аккурат примерно 50 точек)
 - обучаем сетку
 - собираем стату следующие 1 минуту, 2 минуты, три минуты
/show
Tranned by <interval> x 100ms
точность прогноза в первые 20 раз:
time | predict | real | error
     | 0.02%   | %0.2 | 100%    // 0.02 от какого значения?! шит...

реальные цифры подавать - это файл... нужно подавать именно дифы?
типа значения В стакане изменились на % от предыдушего графика, значит они...


точность прогноза за последние 20 раз:
predict | real | error
time | 0.02%   | %0.2 | 100%
...  |         |      | 200%
16:30:02.321


/train 100ms 1000ms - тут еще интересно каков дифф на 100ms... он должен быть больше 0.3%


как мой мозг анализирует картику:
 - смотрю порядка 20 точек на графике и делаю вывод каков тренд и что будет скорее всего дальше
 - потому пока будет брать 50 точек
 - длинна графика будет зависит просто от интервала



 */

/*
TODO
 - собрать данные в json файл, интервалинг
 - нужна нормализация графика, не имеет значения цена реальная, нужны относительные величины... зачем придумывать то за нейронку??

- проще - команда в тг... /train /show-last-1-minute /show-stat


считаем, что предсказание работает в рамках текущего тренда (отрезок в 1 час)
первую минуту собираем данные
тестируем работает ли прогноз еще одну минуту, если больше 60% - то да
запускаем торговлю

через 10 минут паралелльно собираем новые данные, тренеруем НОВУЮ сетку, проверяем
если сетка работает - делаем замену сетки и торгуем дальше, если нет - стопоримся

и так далее. нужно играться с интервалами

*/

/*
Binance:
+44460.00 x 0.318
+44460.30 x 0.001
+44460.61 x 0.007
+44461.45 x 0.000
+44461.78 x 0.002 => 0, 0.002
-44461.99 x 0.884
-44473.08 x 2.300
-44477.05 x 0.029
-44480.16 x 0.720
-44480.82 x 0.250
*/

func main() {
	lines := data()

	const points = 5 // 5 * 6
	const values = 6

	examples := make(Training.Examples, len(lines)-points)

	i := 0
	for {
		slice := lines[i]
		for y := i + 1; y < (i + points); y++ {
			slice = append(slice, lines[y]...)
		}
		if len(lines) > (i+points) {
			examples[i] = Training.Example{
				Input: slice,
				Response: []float64{ lines[i+points][0] },
			}
			i++
		} else {
			break
		}
	}

	Log.Println(examples)

	n := Deep.NewNeural(&Deep.Config{
		/* Input dimensionality */
		Inputs: points * values, // 30
		/* Two hidden layers consisting of two neurons each, and a single output */
		Layout: []int{points * values * 2, points * values * 2, points * values, 10, 1}, // 1000*1000 -> 1
		/* Activation functions: Sigmoid, Tanh, ReLU, Linear */
		Activation: Deep.ActivationSigmoid,
		/* Determines output layer activation & loss function:
		ModeRegression: linear outputs with MSE loss
		ModeMultiClass: softmax output with Cross Entropy loss
		ModeMultiLabel: sigmoid output with Cross Entropy loss
		ModeBinary: sigmoid output with binary CE loss */
		Mode: 4,
		// Mode: Deep.ModeBinary,

		/* Weight initializers: {deep.NewNormal(μ, σ), deep.NewUniform(μ, σ)} */
		Weight: Deep.NewNormal(1.0, 0.0),
		/* Apply bias */
		Bias: true,
	})

	// params: learning rate, momentum, alpha decay, nesterov
	optimizer := Training.NewSGD(0.05, 0.1, 1e-6, true)
	// params: optimizer, verbosity (print stats at every 50th iteration)
	trainer := Training.NewTrainer(optimizer, 100)
	// trainer := Training.NewBatchTrainer(optimizer, 50, 200, 4)

	x1, y2 := examples.Split(0.5)
	trainer.Train(n, x1, y2, 1000) // training, validation, iterations

	Log.Println(examples[5].Input)
	Log.Printf("var1: %v %v", n.Predict(examples[1].Input), examples[1].Response)
	Log.Printf("var1: %v %v", n.Predict(examples[2].Input), examples[2].Response)
	Log.Printf("var1: %v %v", n.Predict(examples[3].Input), examples[3].Response)

}

func data() [][]float64 {
	var lines = [][]float64{
		{46074.83000000, 9391,278.47314000,12809886.29729540,157.54814000,7246781.00833520},
		{46049.93000000, 7657,216.11434000,9940991.21064820,94.60997000,4352001.85129780},
		{46070.00000000, 8323,330.85948000,15222014.58677820,141.75515000,6521035.66470910},
		{46141.78000000, 10364,311.51516000,14318535.31160360,125.53739000,5771478.11039440},
		{45983.00000000, 7406,240.02929000,11025841.02184530,115.51450000,5306033.75249130},
		{46062.26000000, 8934,393.47952000,18110283.88378170,195.12474000,8980529.39062910},
		{46041.59000000, 15448,558.48704000,25614578.21171900,242.84471000,11139810.48690900},
		{46600.00000000, 45301,1997.57928000,92561611.12639600,1094.48243000,50711409.90036790},
		{46585.38000000, 18069,862.19092000,40045381.40969250,438.69917000,20376144.26490890},
		{46615.38000000, 19543,846.60391000,39347258.93122700,397.04135000,18450048.76976560},
		{46678.63000000, 20022,827.37507000,38462113.86055170,449.22998000,20885738.31421900},
		{46750.00000000, 26317, 1110.28498000, 51822885.33955180, 560.77875000, 26174510.39812780},
		{46650.00000000, 19963, 808.86104000, 37648959.06687860, 303.61560000, 14129919.08995460},
		{46464.23000000, 20415, 770.29886000, 35654308.12988220, 337.34039000, 15616308.81092350},
		{46366.35000000, 16917, 663.10566000, 30664247.17508850, 285.47976000, 13207384.05839670},
		{46428.56000000, 13568, 483.96004000, 22431013.37116850, 225.68744000, 10460747.33472840},
		{46577.19000000, 14718, 563.58730000, 26177563.07921040, 254.82446000, 11836899.55636650},
	}

	for x := 0; x < len(lines); x++ {
		lines[x][0] = lines[x][0]/46000
	}

	Log.Println(lines)

	return lines
}
