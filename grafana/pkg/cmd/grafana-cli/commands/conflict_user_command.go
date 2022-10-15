package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/cmd/grafana-cli/logger"
	"github.com/grafana/grafana/pkg/cmd/grafana-cli/utils"
	"github.com/grafana/grafana/pkg/infra/tracing"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/services/sqlstore/db"
	"github.com/grafana/grafana/pkg/services/sqlstore/migrations"
	"github.com/grafana/grafana/pkg/services/user"
	"github.com/grafana/grafana/pkg/services/user/userimpl"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/urfave/cli/v2"
)

func initConflictCfg(cmd *utils.ContextCommandLine) (*setting.Cfg, error) {
	configOptions := strings.Split(cmd.String("configOverrides"), " ")
	configOptions = append(configOptions, cmd.Args().Slice()...)
	cfg, err := setting.NewCfgFromArgs(setting.CommandLineArgs{
		Config:   cmd.ConfigFile(),
		HomePath: cmd.HomePath(),
		Args:     append(configOptions, "cfg:log.level=error"), // tailing arguments have precedence over the options string
	})
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func initializeConflictResolver(cmd *utils.ContextCommandLine, f Formatter, ctx *cli.Context) (*ConflictResolver, error) {
	cfg, err := initConflictCfg(cmd)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", "failed to load configuration", err)
	}
	s, err := getSqlStore(cfg)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", "failed to get to sql", err)
	}
	conflicts, err := GetUsersWithConflictingEmailsOrLogins(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", "failed to get users with conflicting logins", err)
	}
	resolver := ConflictResolver{Users: conflicts}
	resolver.BuildConflictBlocks(conflicts, f)
	return &resolver, nil
}

func getSqlStore(cfg *setting.Cfg) (*sqlstore.SQLStore, error) {
	tracer, err := tracing.ProvideService(cfg)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", "failed to initialize tracer service", err)
	}
	bus := bus.ProvideBus(tracer)
	return sqlstore.ProvideService(cfg, nil, &migrations.OSSMigrations{}, bus, tracer)
}

func runListConflictUsers() func(context *cli.Context) error {
	return func(context *cli.Context) error {
		cmd := &utils.ContextCommandLine{Context: context}
		whiteBold := color.New(color.FgWhite).Add(color.Bold)
		r, err := initializeConflictResolver(cmd, whiteBold.Sprintf, context)
		if err != nil {
			return fmt.Errorf("%v: %w", "failed to initialize conflict resolver", err)
		}
		if len(r.Users) < 1 {
			logger.Info(color.GreenString("No Conflicting users found.\n\n"))
			return nil
		}
		logger.Infof("\n\nShowing conflicts\n\n")
		logger.Infof(r.ToStringPresentation())
		logger.Infof("\n")
		if len(r.DiscardedBlocks) != 0 {
			r.logDiscardedUsers()
		}
		return nil
	}
}

func runGenerateConflictUsersFile() func(context *cli.Context) error {
	return func(context *cli.Context) error {
		cmd := &utils.ContextCommandLine{Context: context}
		r, err := initializeConflictResolver(cmd, fmt.Sprintf, context)
		if err != nil {
			return fmt.Errorf("%v: %w", "failed to initialize conflict resolver", err)
		}
		if len(r.Users) < 1 {
			logger.Info(color.GreenString("No Conflicting users found.\n\n"))
			return nil
		}
		tmpFile, err := generateConflictUsersFile(r)
		if err != nil {
			return fmt.Errorf("generating file return error: %w", err)
		}
		logger.Infof("\n\ngenerated file\n")
		logger.Infof("%s\n\n", tmpFile.Name())
		logger.Infof("once the file is edited and resolved conflicts, you can either validate or ingest the file\n\n")
		if len(r.DiscardedBlocks) != 0 {
			r.logDiscardedUsers()
		}
		return nil
	}
}

func runValidateConflictUsersFile() func(context *cli.Context) error {
	return func(context *cli.Context) error {
		cmd := &utils.ContextCommandLine{Context: context}
		r, err := initializeConflictResolver(cmd, fmt.Sprintf, context)
		if err != nil {
			return fmt.Errorf("%v: %w", "failed to initialize conflict resolver", err)
		}

		// read in the file to validate
		// read in the file to ingest
		arg := cmd.Args().First()
		if arg == "" {
			return errors.New("please specify a absolute path to file to read from")
		}
		b, err := os.ReadFile(filepath.Clean(arg))
		if err != nil {
			return fmt.Errorf("could not read file with error %e", err)
		}
		validErr := getValidConflictUsers(r, b)
		if validErr != nil {
			return fmt.Errorf("could not validate file with error %s", err)
		}
		logger.Info("File validation complete without errors.\n\n File can be used with ingesting command `ingest-file`.\n\n")
		return nil
	}
}

func runIngestConflictUsersFile() func(context *cli.Context) error {
	return func(context *cli.Context) error {
		cmd := &utils.ContextCommandLine{Context: context}
		r, err := initializeConflictResolver(cmd, fmt.Sprintf, context)
		if err != nil {
			return fmt.Errorf("%v: %w", "failed to initialize conflict resolver", err)
		}

		// read in the file to ingest
		arg := cmd.Args().First()
		if arg == "" {
			return errors.New("please specify a absolute path to file to read from")
		}
		b, err := os.ReadFile(filepath.Clean(arg))
		if err != nil {
			return fmt.Errorf("could not read file with error %e", err)
		}
		validErr := getValidConflictUsers(r, b)
		if validErr != nil {
			return fmt.Errorf("could not validate file with error %s", validErr)
		}
		// should we rebuild blocks here?
		// kind of a weird thing maybe?
		if len(r.ValidUsers) == 0 {
			return fmt.Errorf("no users")
		}
		r.showChanges()
		if !confirm("\n\nWe encourage users to create a db backup before running this command. \n Proceed with operation?") {
			return fmt.Errorf("user cancelled")
		}
		err = r.MergeConflictingUsers(context.Context)
		if err != nil {
			return fmt.Errorf("not able to merge with %e", err)
		}
		logger.Info("\n\nconflicts resolved.\n")
		return nil
	}
}

func getDocumentationForFile() string {
	return `# Conflicts File
# This file is generated by the grafana-cli command ` + color.CyanString("grafana-cli admin user-manager conflicts generate-file") + `.
#
# Commands:
# +, keep <user> = keep user
# -, delete <user> = delete user
#
# The fields conflict_email and conflict_login
# indicate that we see a conflict in email and/or login with another user.
# Both these fields can be true.
#
# There needs to be exactly one picked user per conflict block.
#
# The lines can be re-ordered.
#
# If you feel like you want to wait with a specific block,
# delete all lines regarding that conflict block.
#
`
}

func generateConflictUsersFile(r *ConflictResolver) (*os.File, error) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "conflicting_user_*.diff")
	if err != nil {
		return nil, err
	}
	if _, err := tmpFile.Write([]byte(getDocumentationForFile())); err != nil {
		return nil, err
	}
	if _, err := tmpFile.Write([]byte(r.ToStringPresentation())); err != nil {
		return nil, err
	}
	return tmpFile, nil
}

func getValidConflictUsers(r *ConflictResolver, b []byte) error {
	newConflicts := make(ConflictingUsers, 0)
	// need to verify that id or email exists
	previouslySeenIds := map[string]bool{}
	previouslySeenEmails := map[string]bool{}
	previouslySeenLogins := map[string]bool{}
	for _, users := range r.Blocks {
		for _, u := range users {
			previouslySeenIds[strings.ToLower(u.ID)] = true
			previouslySeenEmails[strings.ToLower(u.Email)] = true
			previouslySeenLogins[strings.ToLower(u.Login)] = true
		}
	}

	// tested in https://regex101.com/r/una3zC/1
	diffPattern := `^[+-]`
	// compiling since in a loop
	matchingExpression, err := regexp.Compile(diffPattern)
	if err != nil {
		return fmt.Errorf("unable to compile regex %s: %w", diffPattern, err)
	}
	for _, row := range strings.Split(string(b), "\n") {
		if row == "" {
			// end of file
			break
		}
		// if the row starts with a #, it is a comment
		if row[0] == '#' {
			// comment
			continue
		}
		entryRow := matchingExpression.Match([]byte(row))
		if !entryRow {
			// block row
			// conflict: hej
			continue
		}

		newUser := &ConflictingUser{}
		err := newUser.Marshal(row)
		if err != nil {
			return fmt.Errorf("could not parse the content of the file with error %e", err)
		}
		if newUser.ConflictEmail != "" && !previouslySeenEmails[strings.ToLower(newUser.Email)] {
			return fmt.Errorf("not valid email: %s, email not seen in previous conflicts", newUser.Email)
		}
		if newUser.ConflictLogin != "" && !previouslySeenLogins[strings.ToLower(newUser.Login)] {
			return fmt.Errorf("not valid login: %s, login not seen in previous conflicts", newUser.Login)
		}
		// valid entry
		newConflicts = append(newConflicts, *newUser)
	}
	r.ValidUsers = newConflicts
	r.BuildConflictBlocks(newConflicts, fmt.Sprintf)
	return nil
}

func (r *ConflictResolver) MergeConflictingUsers(ctx context.Context) error {
	for block, users := range r.Blocks {
		if len(users) < 2 {
			return fmt.Errorf("not enough users to perform merge, found %d for id %s, should be at least 2", len(users), block)
		}
		var intoUser user.User
		var intoUserId int64
		var fromUserIds []int64

		// creating a session for each block of users
		// we want to rollback incase something happens during update / delete
		if err := r.Store.WithDbSession(ctx, func(sess *sqlstore.DBSession) error {
			err := sess.Begin()
			if err != nil {
				return fmt.Errorf("could not open a db session: %w", err)
			}
			for _, u := range users {
				if u.Direction == "+" {
					id, err := strconv.ParseInt(u.ID, 10, 64)
					if err != nil {
						return fmt.Errorf("could not convert id in +")
					}
					intoUserId = id
				} else if u.Direction == "-" {
					id, err := strconv.ParseInt(u.ID, 10, 64)
					if err != nil {
						return fmt.Errorf("could not convert id in -")
					}
					fromUserIds = append(fromUserIds, id)
				}
			}
			if _, err := sess.ID(intoUserId).Where(sqlstore.NotServiceAccountFilter(r.Store)).Get(&intoUser); err != nil {
				return fmt.Errorf("could not find intoUser: %w", err)
			}

			for _, fromUserId := range fromUserIds {
				var fromUser user.User
				exists, err := sess.ID(fromUserId).Where(sqlstore.NotServiceAccountFilter(r.Store)).Get(&fromUser)
				if err != nil {
					return fmt.Errorf("could not find fromUser: %w", err)
				}
				if !exists {
					fmt.Printf("user with id %d does not exist, skipping\n", fromUserId)
				}
				// // delete the user
				delErr := r.Store.DeleteUserInSession(ctx, sess, &models.DeleteUserCommand{UserId: fromUserId})
				if delErr != nil {
					return fmt.Errorf("error during deletion of user: %w", delErr)
				}
			}
			commitErr := sess.Commit()
			if commitErr != nil {
				return fmt.Errorf("could not commit operation for useridentification %s: %w", block, commitErr)
			}
			userStore := userimpl.ProvideStore(r.Store, setting.NewCfg())
			updateMainCommand := &user.UpdateUserCommand{
				UserID: intoUser.ID,
				Login:  strings.ToLower(intoUser.Login),
				Email:  strings.ToLower(intoUser.Email),
			}
			updateErr := userStore.Update(ctx, updateMainCommand)
			if updateErr != nil {
				return fmt.Errorf("could not update user: %w", updateErr)
			}

			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

/*
hej@test.com+hej@test.com
all of the permissions, roles and ownership will be transferred to the user.
+ id: 1, email: hej@test.com, login: hej@test.com
these user(s) will be deleted and their permissions transferred.
- id: 2, email: HEJ@TEST.COM, login: HEJ@TEST.COM
- id: 3, email: hej@TEST.com, login: hej@TEST.com
*/
func (r *ConflictResolver) showChanges() {
	if len(r.ValidUsers) == 0 {
		fmt.Println("no changes will take place as we have no valid users.")
		return
	}

	var b strings.Builder
	for block, users := range r.Blocks {
		if _, ok := r.DiscardedBlocks[block]; ok {
			// skip block
			continue
		}

		// looping as we want to can get these out of order (meaning the + and -)
		var mainUser ConflictingUser
		for _, u := range users {
			if u.Direction == "+" {
				mainUser = u
				break
			}
		}
		b.WriteString("Keep the following user.\n")
		b.WriteString(fmt.Sprintf("%s\n", block))
		b.WriteString(fmt.Sprintf("id: %s, email: %s, login: %s\n", mainUser.ID, mainUser.Email, mainUser.Login))
		b.WriteString("\n\n")
		b.WriteString("The following user(s) will be deleted.\n")
		for _, user := range users {
			if user.ID == mainUser.ID {
				continue
			}
			// mergeable users
			b.WriteString(fmt.Sprintf("id: %s, email: %s, login: %s\n", user.ID, user.Email, user.Login))
		}
		b.WriteString("\n\n")
	}
	logger.Info("\n\nChanges that will take place\n\n")
	logger.Infof(b.String())
}

// Formatter make it possible for us to write to terminal and to a file
// with different formats depending on the usecase
type Formatter func(format string, a ...interface{}) string

func shouldDiscardBlock(seenUsersInBlock map[string]string, block string, user ConflictingUser) bool {
	// loop through users to see if we should skip this block
	// we have some more tricky scenarios where we have more than two users that can have conflicts with each other
	// we have made the approach to discard any users that we have seen
	if _, ok := seenUsersInBlock[user.ID]; ok {
		// we have seen the user in different block than the current block
		if seenUsersInBlock[user.ID] != block {
			return true
		}
	}
	seenUsersInBlock[user.ID] = block
	return false
}

// BuildConflictBlocks builds blocks of users where each block is a unique email/login
// NOTE: currently this function assumes that the users are in order of grouping already
func (r *ConflictResolver) BuildConflictBlocks(users ConflictingUsers, f Formatter) {
	discardedBlocks := make(map[string]bool)
	seenUsersToBlock := make(map[string]string)
	blocks := make(map[string]ConflictingUsers)
	for _, user := range users {
		// conflict blocks is how we identify a conflict in the user base.
		var conflictBlock string
		if user.ConflictEmail != "" {
			conflictBlock = f("conflict: %s", strings.ToLower(user.Email))
		} else if user.ConflictLogin != "" {
			conflictBlock = f("conflict: %s", strings.ToLower(user.Login))
		} else if user.ConflictEmail != "" && user.ConflictLogin != "" {
			// both conflicts
			// should not be here unless changed in sql
			conflictBlock = f("conflict: %s%s", strings.ToLower(user.Email), strings.ToLower(user.Login))
		}

		// discard logic
		if shouldDiscardBlock(seenUsersToBlock, conflictBlock, user) {
			discardedBlocks[conflictBlock] = true
		}

		// adding users to blocks
		if _, ok := blocks[conflictBlock]; !ok {
			blocks[conflictBlock] = []ConflictingUser{user}
			continue
		}
		// skip user thats already part of the block
		// since we get duplicate entries
		if contains(blocks[conflictBlock], user) {
			continue
		}
		blocks[conflictBlock] = append(blocks[conflictBlock], user)
	}
	r.Blocks = blocks
	r.DiscardedBlocks = discardedBlocks
}

func contains(cu ConflictingUsers, target ConflictingUser) bool {
	for _, u := range cu {
		if u.ID == target.ID {
			return true
		}
	}
	return false
}

func (r *ConflictResolver) logDiscardedUsers() {
	keys := make([]string, 0, len(r.DiscardedBlocks))
	for block := range r.DiscardedBlocks {
		for _, u := range r.Blocks[block] {
			keys = append(keys, u.ID)
		}
	}
	warn := color.YellowString("Note: We discarded some conflicts that have multiple conflicting types involved.")
	logger.Infof(`
%s

users discarded with more than one conflict:
ids: %s

Solve conflicts and run the command again to see other conflicts.
`, warn, strings.Join(keys, ","))
}

// handling tricky cases::
// if we have seen a user already
// note the conflict of that user
// discard that conflict for next time that the user runs the command

// only present one conflict per user
// go through each conflict email/login
// if any has ids that have already been seen
// discard that conflict
// make note to the user to run again after fixing these conflicts
func (r *ConflictResolver) ToStringPresentation() string {
	/*
		hej@test.com+hej@test.com
		+ id: 1, email: hej@test.com, login: hej@test.com
		- id: 2, email: HEJ@TEST.COM, login: HEJ@TEST.COM
		- id: 3, email: hej@TEST.com, login: hej@TEST.com
	*/
	startOfBlock := make(map[string]bool)
	var b strings.Builder
	for block, users := range r.Blocks {
		if _, ok := r.DiscardedBlocks[block]; ok {
			// skip block
			continue
		}
		for _, user := range users {
			if !startOfBlock[block] {
				b.WriteString(fmt.Sprintf("%s\n", block))
				startOfBlock[block] = true
				b.WriteString(fmt.Sprintf("+ id: %s, email: %s, login: %s, last_seen_at: %s, auth_module: %s, conflict_email: %s, conflict_login: %s\n",
					user.ID,
					user.Email,
					user.Login,
					user.LastSeenAt,
					user.AuthModule,
					user.ConflictEmail,
					user.ConflictLogin,
				))
				continue
			}
			// mergeable users
			b.WriteString(fmt.Sprintf("- id: %s, email: %s, login: %s, last_seen_at: %s, auth_module: %s, conflict_email: %s, conflict_login: %s\n",
				user.ID,
				user.Email,
				user.Login,
				user.LastSeenAt,
				user.AuthModule,
				user.ConflictEmail,
				user.ConflictLogin,
			))
		}
	}
	return b.String()
}

type ConflictResolver struct {
	Store           *sqlstore.SQLStore
	Config          *setting.Cfg
	Users           ConflictingUsers
	ValidUsers      ConflictingUsers
	Blocks          map[string]ConflictingUsers
	DiscardedBlocks map[string]bool
}

type ConflictingUser struct {
	// direction is the +/- which indicates if we should keep or delete the user
	Direction     string `xorm:"direction"`
	ID            string `xorm:"id"`
	Email         string `xorm:"email"`
	Login         string `xorm:"login"`
	LastSeenAt    string `xorm:"last_seen_at"`
	AuthModule    string `xorm:"auth_module"`
	ConflictEmail string `xorm:"conflict_email"`
	ConflictLogin string `xorm:"conflict_login"`
}

type ConflictingUsers []ConflictingUser

func (c *ConflictingUser) Marshal(filerow string) error {
	// example view of the file to ingest
	// +/- id: 1, email: hej, auth_module: LDAP
	trimmedSpaces := strings.ReplaceAll(filerow, " ", "")
	if trimmedSpaces[0] == '+' {
		c.Direction = "+"
	} else if trimmedSpaces[0] == '-' {
		c.Direction = "-"
	} else {
		return fmt.Errorf("unable to get which operation was chosen")
	}
	trimmed := strings.TrimLeft(trimmedSpaces, "+-")
	values := strings.Split(trimmed, ",")

	if len(values) < 3 {
		return fmt.Errorf("expected at least 3 values in entry row")
	}
	// expected fields
	id := strings.Split(values[0], ":")
	email := strings.Split(values[1], ":")
	login := strings.Split(values[2], ":")
	c.ID = id[1]
	c.Email = email[1]
	c.Login = login[1]

	// why trim values, 2022-08-20:19:17:12
	lastSeenAt := strings.TrimPrefix(values[3], "last_seen_at:")
	authModule := strings.Split(values[4], ":")
	if len(authModule) < 2 {
		c.AuthModule = ""
	} else {
		c.AuthModule = authModule[1]
	}
	c.LastSeenAt = lastSeenAt

	// which conflict
	conflictEmail := strings.Split(values[5], ":")
	conflictLogin := strings.Split(values[6], ":")
	if len(conflictEmail) < 2 {
		c.ConflictEmail = ""
	} else {
		c.ConflictEmail = conflictEmail[1]
	}
	if len(conflictLogin) < 2 {
		c.ConflictLogin = ""
	} else {
		c.ConflictLogin = conflictLogin[1]
	}
	return nil
}

func GetUsersWithConflictingEmailsOrLogins(ctx *cli.Context, s *sqlstore.SQLStore) (ConflictingUsers, error) {
	queryUsers := make([]ConflictingUser, 0)
	outerErr := s.WithDbSession(ctx.Context, func(dbSession *sqlstore.DBSession) error {
		rawSQL := conflictingUserEntriesSQL(s)
		err := dbSession.SQL(rawSQL).Find(&queryUsers)
		return err
	})
	if outerErr != nil {
		return queryUsers, outerErr
	}
	return queryUsers, nil
}

// conflictingUserEntriesSQL orders conflicting users by their user_identification
// sorts the users by their useridentification and ids
func conflictingUserEntriesSQL(s *sqlstore.SQLStore) string {
	userDialect := db.DB.GetDialect(s).Quote("user")

	sqlQuery := `
	SELECT DISTINCT
	u1.id,
	u1.email,
	u1.login,
	u1.last_seen_at,
	user_auth.auth_module,
		( SELECT
			'true'
		FROM
			` + userDialect + `
		WHERE (LOWER(u1.email) = LOWER(u2.email)) AND(u1.email != u2.email)) AS conflict_email,
		( SELECT
			'true'
		FROM
			` + userDialect + `
		WHERE (LOWER(u1.login) = LOWER(u2.login) AND(u1.login != u2.login))) AS conflict_login
	FROM
		 ` + userDialect + ` AS u1, ` + userDialect + ` AS u2
	LEFT JOIN user_auth on user_auth.user_id = u1.id
	WHERE (conflict_email IS NOT NULL
		OR conflict_login IS NOT NULL)
		AND (u1.` + notServiceAccount(s) + `)
	ORDER BY conflict_email, conflict_login, u1.id`
	return sqlQuery
}

func notServiceAccount(ss *sqlstore.SQLStore) string {
	return fmt.Sprintf("is_service_account = %s",
		ss.Dialect.BooleanStr(false))
}

// confirm function asks for user input
// returns bool
func confirm(confirmPrompt string) bool {
	var input string
	logger.Infof("%s? [y|n]: ", confirmPrompt)

	_, err := fmt.Scanln(&input)
	if err != nil {
		logger.Infof("could not parse input from user for confirmation")
		return false
	}
	input = strings.ToLower(input)
	if input == "y" || input == "yes" {
		return true
	}
	return false
}
