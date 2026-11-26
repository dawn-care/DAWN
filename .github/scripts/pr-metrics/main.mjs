#!/usr/bin/env node
// Publishes metrics/pull-requests.json: raw facts about every merged PR in this repo (who,
// when, reviews, labels, changed files). It computes nothing; the team site's /metrics page
// turns this into the PR metrics with its own rules.
//
// In Actions, the workflow passes GITHUB_TOKEN. Locally (PowerShell), with the gh CLI signed in:
//   $env:GH_TOKEN = (gh auth token)
//   node .github/scripts/pr-metrics/main.mjs --out ../Internal-Website/public/pull-requests.local.json
// The default --out is metrics/pull-requests.json. Read-only against GitHub.

import { mkdir, writeFile } from 'node:fs/promises'
import { dirname } from 'node:path'
import { createClient, fetchIssueLabels, fetchMergedPrs, fetchPrFiles } from './github.mjs'

const SCHEMA_VERSION = 1

function outPath() {
  const i = process.argv.indexOf('--out')
  return i !== -1 && process.argv[i + 1] ? process.argv[i + 1] : 'metrics/pull-requests.json'
}

/** The issue number a branch name starts with ("481-nighttime-hours" → 481), else null. */
function issueNumberFromBranch(head) {
  const match = /^(\d+)-/.exec(head ?? '')
  return match ? Number(match[1]) : null
}

const labelNames = (connection) => connection.nodes.map((l) => l.name)

async function main() {
  const repo = process.env.GITHUB_REPOSITORY ?? 'dawn-care/DAWN'
  const [owner, name] = repo.split('/')
  const client = createClient(process.env.GITHUB_TOKEN ?? process.env.GH_TOKEN)

  const nodes = await fetchMergedPrs(client, owner, name)
  const issueLabels = new Map()
  const pullRequests = []

  for (const node of nodes) {
    const closingIssues = node.closingIssuesReferences.nodes.map((i) => ({
      number: i.number,
      labels: labelNames(i.labels),
    }))
    // The branch's issue, for PRs that don't link one (labels come from the ticket).
    const branchNumber = closingIssues.length === 0 ? issueNumberFromBranch(node.headRefName) : null
    if (branchNumber !== null && !issueLabels.has(branchNumber)) {
      issueLabels.set(branchNumber, await fetchIssueLabels(client, owner, name, branchNumber))
    }
    const readyEvents = node.timelineItems.nodes.filter((n) => n?.createdAt)

    pullRequests.push({
      number: node.number,
      title: node.title,
      url: node.url,
      author: node.author?.login ?? null,
      authorType: node.author?.__typename ?? null,
      base: node.baseRefName,
      head: node.headRefName,
      createdAt: node.createdAt,
      readyAt: readyEvents.length > 0 ? readyEvents[readyEvents.length - 1].createdAt : null,
      mergedAt: node.mergedAt,
      labels: labelNames(node.labels),
      closingIssues,
      branchIssue:
        branchNumber === null ? null : { number: branchNumber, labels: issueLabels.get(branchNumber) },
      reviews: node.reviews.nodes.map((r) => ({
        author: r.author?.login ?? null,
        authorType: r.author?.__typename ?? null,
        state: r.state,
        submittedAt: r.submittedAt,
      })),
      files: await fetchPrFiles(client, owner, name, node.number),
    })
  }

  const file = { schemaVersion: SCHEMA_VERSION, generatedAt: new Date().toISOString(), repo, pullRequests }
  const out = outPath()
  await mkdir(dirname(out), { recursive: true })
  await writeFile(out, `${JSON.stringify(file)}\n`)
  console.log(`Wrote ${pullRequests.length} merged PRs to ${out}.`)
}

main().catch((error) => {
  console.error(error.message)
  process.exit(1)
})
