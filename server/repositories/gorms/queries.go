package gorms

// SelectLatestDefinition is used to fetch the most recent version of a process definition.
const SelectLatestDefinition = "key = ?"

// OrderLatestDefinition defines the sorting for versions.
const OrderLatestDefinition = "version desc"

// QueryByID is used for finding a record by its primary ID.
const QueryByID = "id = ?"

// QueryByProjectID is used for filtering records by project.
const QueryByProjectID = "project_id = ?"

// UpdateFieldStatus is the column name for status updates.
const UpdateFieldStatus = "status"

// QueryByAssignee is used to filter tasks by their assignee.
const QueryByAssignee = "assignee = ?"

// QueryByInstanceID is used to filter records by process instance ID.
const QueryByInstanceID = "instance_id = ?"

// QueryByStatus is used to filter records by their current status.
const QueryByStatus = "status = ?"

// QueryByCandidateUser is used to check for candidate user in a JSON-stored list.
const QueryByCandidateUser = "candidate_users LIKE ?"

// QueryByCandidateGroup is used to check for candidate group in a JSON-stored list.
const QueryByCandidateGroup = "candidate_groups LIKE ?"

// QueryByDefinitionID is used to filter by definition ID.
const QueryByDefinitionID = "definition_id = ?"

// QueryByPriority is used to filter by task priority.
const QueryByPriority = "priority = ?"

// QueryByUsername is used to filter by user's username.
const QueryByUsername = "username = ?"

// QueryByOrganizationID is used to filter by organization ID.
const QueryByOrganizationID = "organization_id = ?"
