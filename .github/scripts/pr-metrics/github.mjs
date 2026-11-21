// GitHub reads for main.mjs, with Node's built-in fetch (no dependencies). Read-only.

const API = 'https://api.github.com'

export function createClient(token) {
  if (!token) throw new Error('Set GITHUB_TOKEN (or GH_TOKEN) to read pull requests.')
  const headers = {
    Authorization: `Bearer ${token}`,
    Accept: 'application/vnd.github+json',
    'X-GitHub-Api-Version': '2022-11-28',
    'User-Agent': 'dawn-pr-metrics',
  }

  async function request(path, init = {}) {
    const res = await fetch(`${API}${path}`, { ...init, headers: { ...headers, ...init.headers } })
    if (res.status === 404) return null
    if (!res.ok) throw new Error(`GitHub ${init.method ?? 'GET'} ${path} failed: ${res.status}`)
    return res.json()
  }

  async function graphql(query, variables) {
    const body = await request('/graphql', {
      method: 'POST',
      body: JSON.stringify({ query, variables }),
    })
    if (body.errors?.length) {
      throw new Error(`GitHub GraphQL failed: ${body.errors.map((e) => e.message).join('; ')}`)
    }
    return body.data
  }

  return { request, graphql }
}

// author { __typename } tells people from bots: GitHub's own reviewer bots (Copilot's
// "copilot-pull-request-reviewer") don't use the "[bot]" login suffix. 100 reviews and 20
// labels per PR aren't paginated further; no PR here comes close.
const MERGED_PRS = `
  query($owner: String!, $name: String!, $endCursor: String) {
    repository(owner: $owner, name: $name) {
      pullRequests(states: MERGED, first: 50, after: $endCursor, orderBy: {field: CREATED_AT, direction: ASC}) {
        pageInfo { hasNextPage endCursor }
        nodes {
          number title url createdAt mergedAt baseRefName headRefName
          author { __typename login }
          labels(first: 20) { nodes { name } }
          closingIssuesReferences(first: 10) { nodes { number labels(first: 20) { nodes { name } } } }
          reviews(first: 100) { nodes { author { __typename login } state submittedAt } }
          timelineItems(itemTypes: [READY_FOR_REVIEW_EVENT], last: 1) {
            nodes { ... on ReadyForReviewEvent { createdAt } }
          }
        }
      }
    }
  }
`

/** Every merged PR in the repo, as raw GraphQL nodes. */
export async function fetchMergedPrs(client, owner, name) {
  const nodes = []
  let endCursor = null
  do {
    const data = await client.graphql(MERGED_PRS, { owner, name, endCursor })
    const page = data.repository.pullRequests
    nodes.push(...page.nodes)
    endCursor = page.pageInfo.hasNextPage ? page.pageInfo.endCursor : null
  } while (endCursor)
  return nodes
}

/** The PR's changed files ({ path, additions, deletions }), 100 per page. */
export async function fetchPrFiles(client, owner, name, number) {
  const files = []
  for (let page = 1; ; page += 1) {
    const batch = await client.request(
      `/repos/${owner}/${name}/pulls/${number}/files?per_page=100&page=${page}`,
    )
    if (!batch) break
    files.push(
      ...batch.map((f) => ({ path: f.filename, additions: f.additions, deletions: f.deletions })),
    )
    if (batch.length < 100) break
  }
  return files
}

/** An issue's label names ([] if it doesn't exist). */
export async function fetchIssueLabels(client, owner, name, number) {
  const issue = await client.request(`/repos/${owner}/${name}/issues/${number}`)
  return issue ? issue.labels.map((l) => (typeof l === 'string' ? l : l.name)) : []
}
