package errmgr

import (
	"fmt"
	"github.com/olekukonko/errors"
	"testing"
)

func TestMain(m *testing.M) {
	errors.Configure(errors.Config{
		StackDepth:     32,
		ContextSize:    2,
		DisablePooling: false,
		FilterInternal: true,
	})
	Configure(Config{DisableMetrics: false})
	errors.WarmPool(10)
	errors.WarmStackPool(10)
	m.Run()
}

func TestDefine(t *testing.T) {
	ResetCounter("test_tmpl")
	tmpl := Define("test_tmpl", "test error: %s")
	err := tmpl("detail")
	defer err.Free()
	if err.Error() != "test error: detail" {
		t.Errorf("Define() error = %v, want %v", err.Error(), "test error: detail")
	}
	if err.Name() != "test_tmpl" {
		t.Errorf("Define() name = %v, want %v", err.Name(), "test_tmpl")
	}
	if Metrics()["test_tmpl"] != 1 {
		t.Errorf("Metrics()[test_tmpl] = %d, want 1", Metrics()["test_tmpl"])
	}
}

func TestCallable(t *testing.T) {
	ResetCounter("test_call")
	fn := Tracked("test_call", func(args ...interface{}) *errors.Error {
		return errors.Named("test_call").Msgf("called with %v", args[0])
	})
	err := fn("arg1")
	defer err.Free()
	if err.Error() != "called with arg1" {
		t.Errorf("Callable() error = %v, want %v", err.Error(), "called with arg1")
	}
	if Metrics()["test_call"] != 1 {
		t.Errorf("Metrics()[test_call] = %d, want 1", Metrics()["test_call"])
	}
}

func TestCoded(t *testing.T) {
	ResetCounter("test_coded")
	tmpl := Coded("test_coded", "coded error: %s", 400)
	err := tmpl("reason")
	defer err.Free()
	if err.Error() != "coded error: reason" {
		t.Errorf("Coded() error = %v, want %v", err.Error(), "coded error: reason")
	}
	if err.Code() != 400 {
		t.Errorf("Coded() code = %d, want 400", err.Code())
	}
	if Metrics()["test_coded"] != 1 {
		t.Errorf("Metrics()[test_coded] = %d, want 1", Metrics()["test_coded"])
	}
}

func TestMetrics(t *testing.T) {
	Reset()
	ResetCounter("metric1")
	ResetCounter("metric2")
	tmpl1 := Define("metric1", "metric one: %s")
	tmpl2 := Define("metric2", "metric two: %s")

	for i := 0; i < 3; i++ {
		err := tmpl1(fmt.Sprintf("test%d", i))
		err.Free()
	}
	for i := 0; i < 2; i++ {
		err := tmpl2(fmt.Sprintf("test%d", i))
		err.Free()
	}

	metrics := Metrics()
	if len(metrics) != 2 {
		t.Errorf("Metrics() len = %d, want 2", len(metrics))
	}
	if metrics["metric1"] != 3 {
		t.Errorf("Metrics()[metric1] = %d, want 3", metrics["metric1"])
	}
	if metrics["metric2"] != 2 {
		t.Errorf("Metrics()[metric2] = %d, want 2", metrics["metric2"])
	}
}

func TestCountReset(t *testing.T) {
	name := "test_reset"
	ResetCounter(name)
	tmpl := Define(name, "reset test")

	for i := 0; i < 5; i++ {
		err := tmpl("test")
		err.Free()
	}

	err := tmpl("before reset")
	defer err.Free()
	if Metrics()[name] != 6 {
		t.Errorf("Metrics()[%s] before reset = %d, want 6", name, Metrics()[name])
	}

	ResetCounter(name)
	err2 := tmpl("after reset")
	defer err2.Free()
	if Metrics()[name] != 1 {
		t.Errorf("Metrics()[%s] after reset = %d, want 1", name, Metrics()[name])
	}
}
