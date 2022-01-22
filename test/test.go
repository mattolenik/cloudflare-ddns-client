package test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NewTools creates new instances of common testing tools
func NewTools(t *testing.T) (asrt *assert.Assertions, req *require.Assertions, controller *gomock.Controller, cleanup func()) {
	ctrl := gomock.NewController(t)
	return assert.New(t), require.New(t), ctrl, ctrl.Finish
}
