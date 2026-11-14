import assert from 'node:assert/strict'
import { createRequire } from 'node:module'

const require = createRequire(import.meta.url)
const commitizen = require('commitizen')
const config = commitizen.configLoader.load(null, process.cwd())
assert.equal(config.path, 'node_modules/cz-git')
const commitizenRequire = createRequire(require.resolve('commitizen'))
assert.deepEqual(commitizenRequire('glob').sync('package.json'), ['package.json'])
assert.equal(typeof require('cz-git').prompter, 'function')
console.log('Commitizen configuration, glob.sync and cz-git adapter passed')
