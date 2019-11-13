package perceptron

import (
	"fmt"
	"math/rand"
	"mygoml"
	"mygoml/graddesc"

	"gonum.org/v1/gonum/floats"

	"gonum.org/v1/gonum/mat"
)

type Perceptron struct {
	weights mat.Matrix
}

func convert(dataset mygoml.SupervisedDataSet) (mat.Matrix, mat.Matrix) {
	var xdata []float64
	var ydata []float64
	dps := dataset.DataPoints()
	featuresCount := len(dps[0].Features())
	targetCount := len(dps[0].Target())
	for _, dp := range dps {
		xdata = append(xdata, dp.Features()...)
		xdata = append(xdata, 1.0)
		ydata = append(ydata, dp.Target()...)
	}

	xMatrix := mat.NewDense(len(dps), featuresCount+1, xdata)
	yMatrix := mat.NewDense(len(dps), targetCount, ydata)
	return xMatrix.T(), yMatrix.T()
}

func toFloatSlice(m mat.Matrix) []float64 {
	var data []float64
	mt := m.T()
	r, c := mt.Dims()
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			data = append(data, mt.At(i, j))
		}
	}
	return data
}

func copyFloats(x []float64) []float64 {
	n := make([]float64, len(x))
	copy(n, x)
	return n
}

func (p *Perceptron) Train(dataset mygoml.SupervisedDataSet) error {
	X, Y := convert(dataset)
	wr, xcount := X.Dims()
	wc, _ := Y.Dims()
	loss := func(xi, yi []float64) func([]float64) []float64 {
		return func(w []float64) []float64 {
			wMatrix := mat.NewDense(wr, wc, copyFloats(w))
			xVec := mat.NewVecDense(len(xi), copyFloats(xi))
			yVec := mat.NewVecDense(len(yi), copyFloats(yi))
			var temp mat.VecDense
			temp.MulVec(wMatrix.T(), xVec)
			yVec.MulElemVec(yVec, &temp)
			yVec.ScaleVec(-1, yVec)
			result := toFloatSlice(yVec)
			return result
		}
	}
	gradient := func(xi, yi, trueyi []float64) func([]float64) []float64 {
		return func(w []float64) []float64 {
			sameSign := floats.EqualFunc(yi, trueyi, func(a, b float64) bool {
				return a*b > 0
			})
			if sameSign {
				return make([]float64, len(xi))
			}
			xVec := mat.NewVecDense(len(xi), copyFloats(xi))
			wMatrix := mat.NewDense(wr, wc, nil)
			for i, y := range trueyi {
				var wi mat.VecDense
				wi.ScaleVec(-y, xVec)
				wMatrix.SetCol(i, toFloatSlice(&wi))
			}
			return toFloatSlice(wMatrix)
		}
	}

	var wData []float64
	for i := 0; i < wr*wc; i++ {
		wData = append(wData, rand.Float64()/100+0.1)
	}
	W := mat.NewDense(wr, wc, wData)

	var classifiedY mat.Dense
	classifiedY.Mul(W.T(), X)
	epochProvider := &graddesc.StochasticProvider{
		TotalSize: xcount,
		EpochGen: func(i int) graddesc.Function {
			xi := mat.Col(nil, i, X)
			yi := mat.Col(nil, i, &classifiedY)
			trueyi := mat.Col(nil, i, Y)

			return graddesc.Function{
				InputSize: wr * wc,
				Mapper:    loss(xi, yi),
				Gradient:  gradient(xi, yi, trueyi),
			}
		},
		AfterUpdateFunc: func(w []float64) {
			W = mat.NewDense(wr, wc, w)
			classifiedY.Mul(W.T(), X)
		},
		EpochEndFunc: func(w []float64) {},
	}

	op := graddesc.Optimizer{
		EpochProvider: epochProvider,
		LearningRate:  1,
		MaxStep:       100,
		Updater:       &graddesc.BaseUpdater{},
	}

	wData = op.Optimize(wData)
	p.weights = mat.NewDense(wr, wc, wData)
	return nil
}

func (p *Perceptron) Predict(features []float64) ([]float64, error) {
	if r, _ := p.weights.Dims(); len(features) != r-1 {
		msg := fmt.Sprintf("model expects %d features but got %d features", r-1, len(features))
		return nil, mygoml.ErrIncompatibleDataAndModel(msg)
	}

	featureVector := mat.NewVecDense(len(features)+1, append(features, 1))
	var result mat.Dense
	result.Mul(p.weights.T(), featureVector)
	predicted := mat.Col(nil, 0, &result)
	for i := 0; i < len(predicted); i++ {
		if predicted[i] > 0 {
			predicted[i] = 1
		} else if predicted[i] < 0 {
			predicted[i] = -1
		}
	}
	return predicted, nil
}
