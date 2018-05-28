package proj

type ServerCoordinator struct {
    backends []string
    backendClients []*ClientFs
    kc *KeeperClient
}
