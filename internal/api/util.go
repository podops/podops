package api

func canAccessContent(target, ctx, scope string) bool {
	return target == ctx
}
