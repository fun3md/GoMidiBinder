package midi

// Minimal shim of the gomidi API surface used by this project.

type Message []byte

type Sender interface{
    Send(Message) error
}

type Receiver interface{}

type InPort struct{
    n string
}

func (p InPort) Name() string { return p.n }

type OutPort struct{
    n string
}

func (p OutPort) Name() string { return p.n }

func GetInPorts() []InPort { return []InPort{} }
func GetOutPorts() []OutPort { return []OutPort{} }

func FindOutPort(name string) (Sender, error) { return nil, nil }
func FindInPort(name string) (Receiver, error) { return nil, nil }

type ListenOption int

func UseSysEx() ListenOption { return 0 }

// ListenTo is a stub that returns a no-op stop func and no error
func ListenTo(p InPort, cb func(Message, int32), opts ...ListenOption) (func(), error) {
    return func() {}, nil
}

// Listen for a Receiver -- stubbed
func Listen(r Receiver, cb func(Message, int32), opts ...ListenOption) (func(), error) {
    return func() {}, nil
}

func CloseDriver() {}
