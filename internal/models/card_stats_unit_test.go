package models

import "testing"

func TestComputeMedian_OddCount(t *testing.T) {
	result := ComputeMedian([]int{10, 20, 30})
	if result != 20.0 {
		t.Errorf("expected 20.0, got %f", result)
	}
}

func TestComputeMedian_EvenCount(t *testing.T) {
	result := ComputeMedian([]int{10, 20, 30, 40})
	if result != 25.0 {
		t.Errorf("expected 25.0, got %f", result)
	}
}

func TestComputeMedian_Empty(t *testing.T) {
	result := ComputeMedian([]int{})
	if result != 0.0 {
		t.Errorf("expected 0.0, got %f", result)
	}
}

func TestComputeMedian_SingleElement(t *testing.T) {
	result := ComputeMedian([]int{42})
	if result != 42.0 {
		t.Errorf("expected 42.0, got %f", result)
	}
}

func TestComputeWinrate_Normal(t *testing.T) {
	result := ComputeWinrate(3, 1)
	if result != 0.75 {
		t.Errorf("expected 0.75, got %f", result)
	}
}

func TestComputeWinrate_ZeroTotal(t *testing.T) {
	result := ComputeWinrate(0, 0)
	if result != 0.0 {
		t.Errorf("expected 0.0, got %f", result)
	}
}

func TestComputeWinrate_AllWins(t *testing.T) {
	result := ComputeWinrate(5, 0)
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}
