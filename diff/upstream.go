package diff

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteUpstreams() error {
	currentUpstreams, err := sc.currentState.Upstreams.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching upstreams from state")
	}

	for _, upstream := range currentUpstreams {
		n, err := sc.deleteUpstream(upstream)
		if err != nil {
			return err
		}
		if n != nil {
			err = sc.queueEvent(*n)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func (sc *Syncer) deleteUpstream(upstream *state.Upstream) (*Event, error) {
	_, err := sc.targetState.Upstreams.Get(*upstream.ID)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "upstream",
			Obj:  upstream,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up upstream '%v'",
			*upstream.Name)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateUpstreams() error {
	targetUpstreams, err := sc.targetState.Upstreams.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching upstreams from state")
	}

	for _, upstream := range targetUpstreams {
		n, err := sc.createUpdateUpstream(upstream)
		if err != nil {
			return err
		}
		if n != nil {
			err = sc.queueEvent(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sc *Syncer) createUpdateUpstream(upstream *state.Upstream) (*Event,
	error) {
	upstreamCopy := &state.Upstream{Upstream: *upstream.DeepCopy()}
	currentUpstream, err := sc.currentState.Upstreams.Get(*upstream.Name)

	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Create,
			Kind: "upstream",
			Obj:  upstreamCopy,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up upstream %v",
			*upstream.Name)
	}

	// found, check if update needed
	if !currentUpstream.EqualWithOpts(upstreamCopy, false, true) {
		return &Event{
			Op:     crud.Update,
			Kind:   "upstream",
			Obj:    upstreamCopy,
			OldObj: currentUpstream,
		}, nil
	}
	return nil, nil
}
