package cmd

import (
	"github.com/go-errors/errors"
	"github.com/urfave/cli"
	"github.com/ziflex/bumblebee-cli/src/core"
	"github.com/ziflex/bumblebee-cli/src/core/fs"
	"github.com/ziflex/bumblebee-cli/src/system/logging"
	"github.com/ziflex/bumblebee-cli/src/system/storage"
	"github.com/ziflex/bumblebee-cli/src/system/utils"
)

var (
	ERR_ADD_CMD     = errors.New("failed to add new application")
	ERR_GET_ENTRIES = errors.New("failed to get list of registered applications")
	ERR_MISSED_ARGS = errors.New("missed args")
)

type AddCommand struct {
	*cli.Command
	logger   *logging.Logger
	entries  storage.EntryRepository
	settings storage.SettingsRepository
}

func NewAddCommand(logger *logging.Logger, entryRepo storage.EntryRepository, settingsRepo storage.SettingsRepository) *cli.Command {
	return &cli.Command{
		Name:    "add",
		Usage:   "Adds a new application to the app registry",
		Aliases: []string{"a"},
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() == 0 {
				return ERR_MISSED_ARGS
			}

			entries := make([]*core.Entry, 0, ctx.NArg())

			for _, name := range ctx.Args() {
				entries = append(entries, core.NewEntry(name))
			}

			settings, err := settingsRepo.Get()

			if err != nil {
				logger.Error(utils.ErrorStack(err))
				return ERR_ADD_CMD
			}

			transformer := core.NewTransformer(logger, fs.NewDirectory(logger, settings.Directory))

			_, err = transformer.Do(entries, settings.Prefix, false)

			if err != nil {
				logger.Error(utils.ErrorStack(err))
				return ERR_ADD_CMD
			}

			err = entryRepo.CreateMany(entries)

			if err != nil {
				logger.Error(utils.ErrorStack(err))
				return ERR_ADD_CMD
			}

			return nil
		},
	}
}
