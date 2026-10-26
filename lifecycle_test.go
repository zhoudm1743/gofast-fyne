package gofastfyne

import "testing"

// fakeLifecycle 记录注册的回调，供测试手动触发。
type fakeLifecycle struct {
	onStarted, onStopped       func()
	onForeground, onBackground func()
}

func (f *fakeLifecycle) SetOnStarted(fn func())           { f.onStarted = fn }
func (f *fakeLifecycle) SetOnStopped(fn func())           { f.onStopped = fn }
func (f *fakeLifecycle) SetOnEnteredForeground(fn func()) { f.onForeground = fn }
func (f *fakeLifecycle) SetOnExitedForeground(fn func())  { f.onBackground = fn }

func TestWireLifecycle_EmitsAllEvents(t *testing.T) {
	fake := &fakeLifecycle{}
	var got []string
	wireLifecycle(fake, func(name string) { got = append(got, name) })

	fake.onStarted()
	fake.onForeground()
	fake.onBackground()
	fake.onStopped()

	want := []string{EventAppStarted, EventAppForeground, EventAppBackground, EventAppStopped}
	if len(got) != len(want) {
		t.Fatalf("事件序列 = %v, 期望 %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("事件序列 = %v, 期望 %v", got, want)
		}
	}
}
