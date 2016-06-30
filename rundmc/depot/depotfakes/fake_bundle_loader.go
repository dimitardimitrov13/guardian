// This file was generated by counterfeiter
package depotfakes

import (
	"sync"

	"github.com/cloudfoundry-incubator/guardian/rundmc/depot"
	"github.com/cloudfoundry-incubator/guardian/rundmc/goci"
)

type FakeBundleLoader struct {
	LoadStub        func(bundleDir string) (goci.Bndl, error)
	loadMutex       sync.RWMutex
	loadArgsForCall []struct {
		bundleDir string
	}
	loadReturns struct {
		result1 goci.Bndl
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeBundleLoader) Load(bundleDir string) (goci.Bndl, error) {
	fake.loadMutex.Lock()
	fake.loadArgsForCall = append(fake.loadArgsForCall, struct {
		bundleDir string
	}{bundleDir})
	fake.recordInvocation("Load", []interface{}{bundleDir})
	fake.loadMutex.Unlock()
	if fake.LoadStub != nil {
		return fake.LoadStub(bundleDir)
	} else {
		return fake.loadReturns.result1, fake.loadReturns.result2
	}
}

func (fake *FakeBundleLoader) LoadCallCount() int {
	fake.loadMutex.RLock()
	defer fake.loadMutex.RUnlock()
	return len(fake.loadArgsForCall)
}

func (fake *FakeBundleLoader) LoadArgsForCall(i int) string {
	fake.loadMutex.RLock()
	defer fake.loadMutex.RUnlock()
	return fake.loadArgsForCall[i].bundleDir
}

func (fake *FakeBundleLoader) LoadReturns(result1 goci.Bndl, result2 error) {
	fake.LoadStub = nil
	fake.loadReturns = struct {
		result1 goci.Bndl
		result2 error
	}{result1, result2}
}

func (fake *FakeBundleLoader) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.loadMutex.RLock()
	defer fake.loadMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeBundleLoader) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ depot.BundleLoader = new(FakeBundleLoader)
