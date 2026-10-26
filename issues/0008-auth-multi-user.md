---
labels: [wayfinder:grilling]
assignee:
---

# 0008 · Authentication and multi-user model

## Question

What is the real auth model after the skeleton's development mode (trusted `X-User-Id` header): agent enrollment and identity, user accounts vs API tokens, and what the API shape must reserve so auth can harden without breaking clients? Non-blocking for the skeleton; resolve mid-map so the protocol (0006) and API surface don't paint themselves into a corner.
