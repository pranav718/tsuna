package p2p

type Sender interface {
	Send(msgType MsgType, payload any) error
	Recv() <-chan *Envelope
	Close()
}
