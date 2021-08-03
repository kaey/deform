package main

func (terp *Interp) initBuiltinKinds() {
	terp.Kinds["object"] = Kind{}
	terp.Kinds["direction"] = Kind{Kind: "object"}
	terp.Kinds["room"] = Kind{Kind: "object"}
	terp.Kinds["region"] = Kind{Kind: "object"}
	terp.Kinds["thing"] = Kind{Kind: "object"}
	terp.Kinds["door"] = Kind{Kind: "thing"}
	terp.Kinds["container"] = Kind{Kind: "thing"}
	terp.Kinds["player's holdall"] = Kind{Kind: "container"}
	terp.Kinds["supporter"] = Kind{Kind: "thing"}
	terp.Kinds["backdrop"] = Kind{Kind: "thing"}
	terp.Kinds["device"] = Kind{Kind: "thing"}
	terp.Kinds["person"] = Kind{Kind: "thing"}
	terp.Kinds["man"] = Kind{Kind: "person"}
	terp.Kinds["woman"] = Kind{Kind: "person"}
	terp.Kinds["animal"] = Kind{Kind: "person"}
}
