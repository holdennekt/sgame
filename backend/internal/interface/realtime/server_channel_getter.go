package realtime

type ServerChannelGetter interface {
	Get(name string) any
}
