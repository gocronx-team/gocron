# Contributing to gocron

Thanks for your interest in contributing! 🎉

## Workflow

1. **Fork** the repository and **clone** your fork:

   ```bash
   git clone https://github.com/YOUR_USERNAME/gocron.git
   cd gocron
   ```

2. **Install dependencies** (also installs the git hooks):

   ```bash
   pnpm install
   pnpm run prepare
   ```

3. **Create a feature branch:**

   ```bash
   git checkout -b feature/your-feature-name
   ```

4. **Make your changes and commit** — see [Commit messages](#commit-messages) below.

5. **Push and open a Pull Request:**

   ```bash
   git push origin feature/your-feature-name
   ```

## Commit messages

This project enforces [Conventional Commits](https://www.conventionalcommits.org/) through a
[commitlint](https://github.com/conventional-changelog/commitlint) git hook, so a plain
`git commit` with a free-form message will be rejected. Use the interactive tool
([commitizen](https://github.com/commitizen/cz-cli) + [cz-git](https://cz-git.qbb.sh/)) instead:

```bash
git add .
pnpm run commit
```

It guides you through producing messages such as:

- `feat(task): add task dependency configuration`
- `fix(api): fix task status update issue`
- `docs: update API documentation`

Avoid raw `@` mentions in commit subjects and pull request titles: GitHub may
notify an unrelated account. For cron expressions such as the reboot trigger,
use a descriptive word in the title and put the exact syntax in the body.
