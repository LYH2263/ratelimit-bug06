package syncstate

func Match(cur, expect State) bool {
	if expect.Version == 0 {
		return cur.Version == 0
	}
	return cur.Version == expect.Version
}
