package mux

type Conn struct {
  id string
  muxer *Muxer
}

func (conn *Conn) Write(p []byte) (n int, err error) {
  return conn.muxer.Write(conn.id, p)
}

func (conn *Conn) Read(p []byte) (n int, err error) {
  return conn.muxer.Read(conn.id, p)
}

func (conn *Conn) Close() error {
  return conn.muxer.CloseConn(conn.id)
}

func (conn *Conn) String() string {
  return conn.id
}
