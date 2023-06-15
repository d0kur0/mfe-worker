package configMap

var ConfigTemplate = ConfigMap{
	HttpBaseUrl: "[base http url: ex: http://localhost:3433]",
	DBPath:      "[path for save sqlite db file, ex: mf_worker.db]",
	GitlabUrl:   "[base gitlab instance url]",
	GitlabToken: "[gitlab token with access to read projects]",
	StoragePath: "[path to dir for store build assets]",
	Projects: []Project{{
		Branches:      []string{"[branches white list or empty array for pass all names]"},
		ProjectID:     "[project id of gitlab]",
		DistFiles:     []string{"[files what need to save after build and share]", "dist/app.js", "dist/app.css"},
		ProjectName:   "[project name (any value, not gitlab name)]",
		BuildCommands: []string{"[commands for build project after clone]", "npm run prebuild", "npm run build"},
	}},
}
