package model

import (
	"fmt"
	"testing"
)

func BenchmarkRegistryAdd(b *testing.B) {
	r := NewRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		job := NewJob(fmt.Sprintf("job-%v", b.N), fmt.Sprintf("echo"))
		r.Add(job)
	}
}

func BenchmarkRegistryCleanUp(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := NewRegistry()
		for ii := 0; ii < 100; ii++ {
			job := NewJob(fmt.Sprintf("job-%v", b.N), fmt.Sprintf("echo"))
			r.Add(job)
			r.Cleanup()
		}
	}
}

func TestRegistryAddNoDuplicateJob(t *testing.T) {
	r := NewRegistry()
	for ii := 0; ii < 100; ii++ {
		job := NewJob(fmt.Sprintf("job-%v", ii), fmt.Sprintf("echo"))
		if !r.Add(job) {
			t.Errorf("Expect to add job")
		}
		for j := 0; j < 10; j++ {
			if r.Add(job) {
				t.Errorf("Expect not to add job")

			}
		}

	}
}

func TestRegistryLen(t *testing.T) {
	r := NewRegistry()
	num := 100
	for ii := 0; ii < num; ii++ {
		job := NewJob(fmt.Sprintf("job-%v", ii), fmt.Sprintf("echo"))
		if !r.Add(job) {
			t.Errorf("Expect to add job")
		}
	}
	if r.Len() != num {
		t.Errorf("Expect %v got length %v", num, r.Len())
	}
}

func TestRegistryDelete(t *testing.T) {
	r := NewRegistry()
	num := 100
	for ii := 0; ii < num; ii++ {
		job := NewJob(fmt.Sprintf("job-%v", ii), fmt.Sprintf("echo"))
		if !r.Add(job) {
			t.Errorf("Expect to add job")
		}
		if !r.Delete(job.Id) {
			t.Errorf("Expect to delete job")
		}
		if r.Delete(job.Id) {
			t.Errorf("Expect the job to be already deleted")
		}

	}
	if r.Len() != 0 {
		t.Errorf("Expect %v got length %v", num, r.Len())
	}
}
