package main

type Actor = func(*HookData) error

var ActorGitSync Actor = func(hookData *HookData) error {
	return nil
}
