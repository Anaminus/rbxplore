package cmd

import (
	"errors"
	"github.com/anaminus/rbxplore/action"
	"github.com/robloxapi/rbxfile"
)

////////////////

func AddRootInstance(root *rbxfile.Root, inst *rbxfile.Instance) action.Action {
	return &actionAddRoot{root, inst}
}

type actionAddRoot struct {
	root     *rbxfile.Root
	instance *rbxfile.Instance
}

func (a *actionAddRoot) Setup() error {
	return nil
}

func (a *actionAddRoot) Forward() error {
	a.root.Instances = append(a.root.Instances, a.instance)
	return nil
}

func (a *actionAddRoot) Backward() error {
	a.root.Instances[len(a.root.Instances)-1] = nil
	a.root.Instances = a.root.Instances[:len(a.root.Instances)-1]
	return nil
}

////////////////

func RemoveRootInstance(root *rbxfile.Root, i int) action.Action {
	return &actionRemoveRoot{root: root, index: i}
}

type actionRemoveRoot struct {
	root     *rbxfile.Root
	index    int
	instance *rbxfile.Instance
}

func (a *actionRemoveRoot) Setup() error {
	if a.index < 0 || a.index >= len(a.root.Instances) {
		return errors.New("index out of range")
	}
	a.instance = a.root.Instances[a.index]
	return nil
}

func (a *actionRemoveRoot) Forward() error {
	copy(a.root.Instances[a.index:], a.root.Instances[a.index+1:])
	a.root.Instances[len(a.root.Instances)-1] = nil
	a.root.Instances = a.root.Instances[:len(a.root.Instances)-1]
	return nil
}

func (a *actionRemoveRoot) Backward() error {
	a.root.Instances = append(a.root.Instances, nil)
	copy(a.root.Instances[a.index+1:], a.root.Instances[a.index:])
	a.root.Instances[a.index] = a.instance
	return nil
}

////////////////

func SetClassName(inst *rbxfile.Instance, className string) action.Action {
	return &actionSetClassName{instance: inst, newClassName: className}
}

type actionSetClassName struct {
	instance     *rbxfile.Instance
	newClassName string
	oldClassName string
}

func (a *actionSetClassName) Setup() error {
	a.oldClassName = a.instance.ClassName
	return nil
}

func (a *actionSetClassName) Forward() error {
	a.instance.ClassName = a.newClassName
	return nil
}

func (a *actionSetClassName) Backward() error {
	a.instance.ClassName = a.oldClassName
	return nil
}

////////////////

func SetReference(inst *rbxfile.Instance, ref []byte) action.Action {
	c := make([]byte, len(ref))
	copy(c, ref)
	return &actionSetReference{instance: inst, newRef: c}
}

type actionSetReference struct {
	instance *rbxfile.Instance
	newRef   string
	oldRef   string
}

func (a *actionSetReference) Setup() error {
	a.oldRef = a.instance.Reference
	return nil
}

func (a *actionSetReference) Forward() error {
	a.instance.Reference = a.newRef
	return nil
}

func (a *actionSetReference) Backward() error {
	a.instance.Reference = a.oldRef
	return nil
}

////////////////

func SetParent(inst, parent *rbxfile.Instance) action.Action {
	return &actionSetParent{instance: inst, newParent: parent}
}

type actionSetParent struct {
	instance  *rbxfile.Instance
	newParent *rbxfile.Instance
	oldParent *rbxfile.Instance
	oldIndex  int
}

func (a *actionSetParent) Setup() error {
	a.oldParent = a.instance.Parent()
	if a.oldParent != nil {
		for i, c := range a.oldParent.Children {
			if c == a.instance {
				a.oldIndex = i
				return nil
			}
		}
		return errors.New("instance is not a child of parent")
	}
	return nil
}

func (a *actionSetParent) Forward() error {
	return a.instance.SetParent(a.newParent)
}

func (a *actionSetParent) Backward() error {
	if a.oldParent == nil {
		return a.instance.SetParent(nil)
	} else {
		return a.oldParent.AddChildAt(a.oldIndex, a.instance)
	}
}

////////////////

func SetIsService(inst *rbxfile.Instance, isService bool) action.Action {
	return &actionSetIsService{instance: inst, newIsService: isService}
}

type actionSetIsService struct {
	instance     *rbxfile.Instance
	newIsService bool
	oldIsService bool
}

func (a *actionSetIsService) Setup() error {
	a.oldIsService = a.instance.IsService
	return nil
}

func (a *actionSetIsService) Forward() error {
	a.instance.IsService = a.newIsService
	return nil
}

func (a *actionSetIsService) Backward() error {
	a.instance.IsService = a.oldIsService
	return nil
}

////////////////

func SetProperty(inst *rbxfile.Instance, prop string, value rbxfile.Value) action.Action {
	return &actionSetProperty{instance: inst, prop: prop, newValue: value}
}

type actionSetProperty struct {
	instance *rbxfile.Instance
	prop     string
	newValue rbxfile.Value
	oldValue rbxfile.Value
}

func (a *actionSetProperty) Setup() error {
	if old, ok := a.instance.Properties[a.prop]; ok {
		a.oldValue = old.Copy()
	}
	return nil
}

func (a *actionSetProperty) Forward() error {
	a.instance.Properties[a.prop] = a.newValue
	return nil
}

func (a *actionSetProperty) Backward() error {
	if a.oldValue == nil {
		delete(a.instance.Properties, a.prop)
	} else {
		a.instance.Properties[a.prop] = a.oldValue
	}
	return nil
}

////////////////
