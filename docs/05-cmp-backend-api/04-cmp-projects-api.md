# Projects & RBAC API Contract

The Projects API manages organizational boundaries (Projects) and handles team member allocation (RBAC) by directly synchronizing roles with Keycloak groups and Vault policies.

---

## 1. List User Projects

Retrieves the list of projects the authenticated user belongs to. Role mappings are dynamically fetched from Keycloak to prevent stale token claims.

* **HTTP Method**: `GET`
* **Path**: `/api/projects`
* **Authorization**: Requires a valid Bearer JWT.
* **Response (`200 OK`)**:

```json
[
  {
    "name": "alpha",
    "role": "owner"
  },
  {
    "name": "beta",
    "role": "member"
  }
]
```
*Note: Roles can be `"owner"`, `"admin"`, or `"member"`.*

---

## 2. Create Project (Bootstrap)

Launches an asynchronous background task to run the `k3s-project-bootstrap` Terraform module. This creates Keycloak groups, Vault namespaces, and ArgoCD AppProjects.

* **HTTP Method**: `POST`
* **Path**: `/api/projects`
* **Content-Type**: `application/json`
* **Request Payload (`ProjectCreate`)**:

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `project_name` | String | Yes | Lowercase, kebab-case project identifier (e.g., `alpha-team`). |

### Example Request
```bash
curl -X POST "https://cmp.3istor.com/api/projects" \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "project_name": "alpha"
  }'
```

**Expected Response (`202 Accepted`)**:
```json
{
  "message": "Project 'alpha' bootstrap started. Keycloak groups, Vault policies, and ArgoCD AppProject will be created shortly. You will be added as project admin.",
  "project_name": "alpha",
  "status": "bootstrapping"
}
```

---

## 3. List Project Members

Returns all members allocated to the project. Performs parallel Admin API queries to Keycloak.

* **HTTP Method**: `GET`
* **Path**: `/api/projects/{project_name}/members`
* **Required Roles**: `project-{project_name}-members` | `project-{project_name}-admins`
* **Response (`200 OK`)**:

```json
{
  "project_name": "alpha",
  "members": [
    {
      "username": "brian.perret",
      "email": "brian.perret@epita.fr",
      "first_name": "Brian",
      "last_name": "Perret",
      "role": "owner"
    },
    {
      "username": "joe.bejjani",
      "email": "joe.bejjani@epita.fr",
      "first_name": "Joe",
      "last_name": "Bejjani",
      "role": "admin"
    },
    {
      "username": "raphael.ye",
      "email": "raphael.ye@epita.fr",
      "first_name": "Raphael",
      "last_name": "Ye",
      "role": "member"
    }
  ]
}
```

---

## 4. Add Project Member

Adds a Keycloak user to either the admin or member group associated with the project.

* **HTTP Method**: `POST`
* **Path**: `/api/projects/{project_name}/members`
* **Required Roles**: `project-{project_name}-admins` (Only admins can add members)
* **Request Payload**:

```json
{
  "username": "raphael.ye",
  "role": "member"
}
```

### Example Request
```bash
curl -X POST "https://cmp.3istor.com/api/projects/alpha/members" \
  -H "Authorization: Bearer eyJhbGci..." \
  -H "Content-Type: application/json" \
  -d '{
    "username": "raphael.ye",
    "role": "member"
  }'
```

**Expected Response (`201 Created`)**:
```json
{
  "message": "User 'raphael.ye' added to project 'alpha' with role 'member'.",
  "project_name": "alpha",
  "username": "raphael.ye",
  "role": "member"
}
```

---

## 5. Remove Project Member

Removes a user from both admin and member Keycloak groups for the given project.

* **HTTP Method**: `DELETE`
* **Path**: `/api/projects/{project_name}/members/{username}`
* **Required Roles**: `project-{project_name}-admins` (Only admins can remove members)
* **Response (`204 No Content`)**: No body returned on success.

```bash
curl -X DELETE "https://cmp.3istor.com/api/projects/alpha/members/raphael.ye" \
  -H "Authorization: Bearer eyJhbGci..."
```

---

## 6. Search Users (Autocomplete)

Prefix search against Keycloak users by username, email, first name, or last name. Used to populate the membership auto-discovery dropdown on the frontend.

* **HTTP Method**: `GET`
* **Path**: `/api/projects/users/search`
* **Query Parameters**:
  * `q`: Search string (Minimum 2 characters required).
* **Response (`200 OK`)**:

```json
[
  {
    "username": "raphael.ye",
    "email": "raphael.ye@epita.fr",
    "first_name": "Raphael",
    "last_name": "Ye"
  }
]
```

---

## 7. Delete Project

Wipes out the project's Keycloak groups, S3 state configurations, and ownership records.

* **HTTP Method**: `DELETE`
* **Path**: `/api/projects/{project_name}`
* **Required Roles**: `project-{project_name}-admins`
* **Safety Lock**: The API returns an error if any non-deleted applications exist in the project.

### Example Request
```bash
curl -X DELETE "https://cmp.3istor.com/api/projects/alpha" \
  -H "Authorization: Bearer eyJhbGci..."
```

**Error Response (`400 Bad Request` - safety lock triggered)**:
```json
{
  "detail": "Cannot delete project 'alpha': it has 2 active application(s). Delete all applications first."
}
```

**Expected Response (`204 No Content`)**: No body returned on successful deletion.

---
**Next Step**: Continue to [Catalog API Contract](05-cmp-catalog-api.md) (or return to the [Project Overview](../index.md)).
