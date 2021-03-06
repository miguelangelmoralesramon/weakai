package svm

import "math"

// A SubgradientSolver solves Problems using sub-gradient descent.
//
// This is only guaranteed to be effective for linear kernels, since such kernels yield a convex
// function to optimize.
type SubgradientSolver struct {
	// Tradeoff specifies how important it is to minimize the magnitude of the normal vector versus
	// finding a good separation of samples.
	// In other words, it determines how important a wide margin is.
	// For linearly separable data, you should use a small (but non-zero) Tradeoff value.
	Tradeoff float64

	// Steps indicates how many descents the solver should make before returning its solution.
	// Increasing the number of steps will increase the accuracy, but by decreasing amounts.
	Steps int

	// StepSize is a number between 0 and 1 which determines how much of the gradient should be
	// added to the current solution at each step.
	// Values closer to 0 will result in better accuracy, while values closer to 1 will cause the
	// solver to approach the solution in fewer steps.
	StepSize float64
}

func (s *SubgradientSolver) Solve(p *Problem) *LinearClassifier {
	args := softMarginArgs{
		normal: make([]float64, len(p.Positives[0].V)),
	}

	for i := 0; i < s.Steps; i++ {
		args = s.descend(p, args)
	}

	return &LinearClassifier{
		HyperplaneNormal: Sample{V: args.normal},
		Threshold:        args.threshold,
		Kernel:           p.Kernel,
	}
}

func (s *SubgradientSolver) descend(p *Problem, args softMarginArgs) softMarginArgs {
	res := args
	res.normal = make([]float64, len(args.normal))
	copy(res.normal, args.normal)

	res.threshold -= s.thresholdPartial(p, args) * s.StepSize
	for i := range res.normal {
		res.normal[i] -= s.normalPartial(p, args, i) * s.StepSize
	}

	return res
}

// thresholdPartial approximates the partial differential of the soft-margin function with respect
// to the threshold argument.
func (s *SubgradientSolver) thresholdPartial(p *Problem, args softMarginArgs) float64 {
	// TODO: figure out a good "differential" value.
	differential := 1.0 / 10000.0

	tempArgs := args
	tempArgs.threshold += differential
	return (s.softMarginFunction(p, tempArgs) - s.softMarginFunction(p, args)) / differential
}

// normalPartial approximates the partial differential of the soft-margin function with respect to
// a component of the normal vector.
func (s *SubgradientSolver) normalPartial(p *Problem, args softMarginArgs, comp int) float64 {
	// TODO: figure out a good "differential" value.
	differential := 1.0 / 10000.0

	tempArgs := args
	tempArgs.normal = make([]float64, len(args.normal))
	copy(tempArgs.normal, args.normal)
	tempArgs.normal[comp] += differential
	return (s.softMarginFunction(p, tempArgs) - s.softMarginFunction(p, args)) / differential
}

func (s *SubgradientSolver) softMarginFunction(p *Problem, args softMarginArgs) float64 {
	normalSample := Sample{V: args.normal}

	var matchSum float64
	for _, positive := range p.Positives {
		errorMargin := math.Max(0, 1-(p.Kernel(normalSample, positive)+args.threshold))
		matchSum += errorMargin
	}
	for _, negative := range p.Negatives {
		errorMargin := math.Max(0, 1+(p.Kernel(normalSample, negative)+args.threshold))
		matchSum += errorMargin
	}
	return matchSum + s.Tradeoff*p.Kernel(normalSample, normalSample)
}

type softMarginArgs struct {
	normal    []float64
	threshold float64
}
