package parser

var descriptions = map[string]*descriptionInfo{
	"du":                                    newDescriptionInfo("Document Understanding", "Document Understanding uses a combination of robotic process automation (RPA) and AI to automatically process your documents."),
	"du classification":                     newDescriptionInfo("Sorting documents into categories using automation", "Document Classification is a component that helps in identifying what types of files the robot is processing."),
	"du digitization":                       newDescriptionInfo("Extracting Document Object Model (DOM) and text", "Digitization is the process of obtaining machine readable text from a given incoming file, so that a robot can then understand its contents and act upon them."),
	"du discovery":                          newDescriptionInfo("Discover projects and resources", "Using the Discovery Service for retrieving information about the newly created project and its resources."),
	"du extraction":                         newDescriptionInfo("Extracting and processing data from documents", "Data Extraction is a component that helps in identifying very specific information that you are interested in, from your document types."),
	"du validation":                         newDescriptionInfo("Ensuring accuracy of extracted data", "Validation is an optional step which refers to a human review step, in which knowledge workers can review the results and correct them when necessary."),
	"identity":                              newDescriptionInfo("Identity Server", "Service that offers centralized authentication and access control across UiPath products."),
	"identity audit-query":                  newDescriptionInfo("Inspect audit events", "As an organization admin, use these endpoints to list and download audit events."),
	"identity group":                        newDescriptionInfo("Manage groups in your organization", "Use these endpoints to manage groups in your organization."),
	"identity message-template":             newDescriptionInfo("Create message Templates", "Manage your message templates."),
	"identity robot-account":                newDescriptionInfo("Manage robot accounts", "Use these endpoints to manage robot accounts."),
	"identity setting":                      newDescriptionInfo("Update application settings", "Manage your application settings."),
	"identity token":                        newDescriptionInfo("Access tokens", "Endpoint to retrieve short-lived access tokens."),
	"identity user":                         newDescriptionInfo("Manage local users", "Use these endpoints to manage local users."),
	"identity user-login-attempt":           newDescriptionInfo("Inspect user login attempts", "Review login attempts of your users."),
	"orchestrator":                          newDescriptionInfo("UiPath Orchestrator", "Orchestrator gives you the power you need to provision, deploy, trigger, monitor, measure, and track the work of attended and unattended robots."),
	"orchestrator alerts":                   newDescriptionInfo("Notifications for system events", "Notifications related to robots, queue items, triggers, and more."),
	"orchestrator app-tasks":                newDescriptionInfo("Tasks executed within Orchestrator apps", "App Tasks are activities and assignments that need human intervention in a robotic process automation (RPA) workflow."),
	"orchestrator assets":                   newDescriptionInfo("Shared, reusable workflow values", "Assets usually represent shared variables or credentials that can be used in different automation projects. They allow you to store specific information so that the robots can easily access it."),
	"orchestrator audit-logs":               newDescriptionInfo("Recorded activity for audit purposes", "Logs help with debugging issues, increasing security and performance, or reporting trends."),
	"orchestrator buckets":                  newDescriptionInfo("Storage for specific workflow data", "Buckets provide a per-folder storage solution for RPA developers to leverage in creating automation projects."),
	"orchestrator business-rules":           newDescriptionInfo("Policies governing system processes", "Business rules are predefined instructions or conditions that determine how a process or task should be carried out."),
	"orchestrator calendars":                newDescriptionInfo("Scheduling automation tasks", "Calendars allow users to fine-tune schedules for jobs in your automations."),
	"orchestrator credential-stores":        newDescriptionInfo("Secure storage for access credentials", "Secure location where you can store sensitive data."),
	"orchestrator directory-service":        newDescriptionInfo("Managing active directory", "Integration of Orchestrator with an organization's directory service to manage users and group membership."),
	"orchestrator environments":             newDescriptionInfo("Grouping Robots for deployments", "Environments represent a logical grouping of robots that have common characteristics or purposes."),
	"orchestrator execution-media":          newDescriptionInfo("Media from a process run", "Storing screenshots of RPA executions, so that developers, administrators, or process stakeholders could see what actions a robot took during execution."),
	"orchestrator exports":                  newDescriptionInfo("Data exported from Orchestrator", "Extract and download data into a file that you can use for reporting, analysis, or for backup and archival purposes."),
	"orchestrator folders":                  newDescriptionInfo("Organizational structure for resources", "Folders are organizational units that help manage and organize resources such as robots, processes, queues, assets, and more."),
	"orchestrator folders-navigation":       newDescriptionInfo("Hierarchical traversal through folders", "Allows users to manage and organize resources effectively, including robots, processes, queues, and assets."),
	"orchestrator generic-tasks":            newDescriptionInfo("Basic, undefined task types", "Manage basic unit of work assigned by a process or a robot."),
	"orchestrator job-triggers":             newDescriptionInfo("Initiators of jobs", "Triggers are used to schedule the execution of processes."),
	"orchestrator jobs":                     newDescriptionInfo("Specific instances of process execution", "Job represent an execution of a process on the UiPath robots."),
	"orchestrator libraries":                newDescriptionInfo("Repositories for reusable components", "Manage reusable components used in automation workflows."),
	"orchestrator licenses-named-user":      newDescriptionInfo("User-specific licenses", "Licenses granted to a specific user in the system, typically identified by their username."),
	"orchestrator licenses-runtime":         newDescriptionInfo("Licenses for robot runtime capacity", "A runtime license refers to a licensing model where the license is granted to an unattended robot per run."),
	"orchestrator licensing":                newDescriptionInfo("License units", "APIs to aquire and release license units."),
	"orchestrator logs":                     newDescriptionInfo("Recorded activities and events", "Records of the activities and events that occur during the execution of a process."),
	"orchestrator machines":                 newDescriptionInfo("Managed hosts for robots", "Physical or virtual machine (computer) where to deploy a UiPath robot for executing automation processes."),
	"orchestrator maintenance":              newDescriptionInfo("Maintenance mode", "Maintenance mode provides a simplified solution for stopping all Orchestrator activity."),
	"orchestrator organization-units":       newDescriptionInfo("Divisions for resource management", "Entities that can have their own separate robots, processes, assets, queues, etc.."),
	"orchestrator package-feeds":            newDescriptionInfo("Packages management", "Package feeds represent the sources where the automation packages or processes developed in UiPath Studio are stored so they can be distributed and run on robots."),
	"orchestrator permissions":              newDescriptionInfo("Access rights", "Permissions define the level of access a user has within the system. They are used to control which features and resources a user can view, edit, or manage."),
	"orchestrator personal-workspaces":      newDescriptionInfo("Individual work areas", "Users can have personal workspaces where they can publish, test, and run their processes. Personal workspaces serve as a staging or development area for users to refine their automation before moving it to a shared workspace or production."),
	"orchestrator process-schedules":        newDescriptionInfo("Planned schedules for process execution", "Process schedules allow you to automate when your processes or jobs are run, without having to manually start them each time."),
	"orchestrator processes":                newDescriptionInfo("Workflows for execution", "Processes refer to the actual automations or workflows developed using UiPath Studio that are to be executed by the robots."),
	"orchestrator queue-definitions":        newDescriptionInfo("Descriptions for types of queues", "Queue definitions are used to define and configure the queues that are used in robotic Process Automation (RPA) workflows."),
	"orchestrator queue-item-comments":      newDescriptionInfo("Notes for queue items", "Add comments or notes to a specific queue item."),
	"orchestrator queue-item-events":        newDescriptionInfo("Events related to queue items", "Events are triggered when the status of a queue item changes, such as when it is added, started, failed, successful etc."),
	"orchestrator queue-items":              newDescriptionInfo("Items to be processed by robots", "Queue items are free form data to be processed by robots in automation workflows."),
	"orchestrator queue-processing-records": newDescriptionInfo("Processing records of queue items", "Information about processing queue items, including status updates and reasons for failure."),
	"orchestrator queue-retention":          newDescriptionInfo("Rules for storing queue data", "Queue retention defines how long the processed queue item data is stored before it is deleted."),
	"orchestrator queues":                   newDescriptionInfo("Storage for multiple work items", "Queues are used to store multiple items that need to be processed by robots."),
	"orchestrator release-retention":        newDescriptionInfo("Rules for retaining release information", "Release retention define the rules for how long the information about a release is retained in Orchestrator's storage."),
	"orchestrator releases":                 newDescriptionInfo("Packages prepared for robot execution", "Releases represent the processes or packages that are prepared for execution on robots."),
	"orchestrator robot-logs":               newDescriptionInfo("Records of robot events", "robot Logs are records of the different actions and events tracked during a robot's operation."),
	"orchestrator robots":                   newDescriptionInfo("Entities executing automation processes", "robots are the entities that execute the automation processes."),
	"orchestrator roles":                    newDescriptionInfo("User permissions management system", "Roles allow users to assign and manage different permissions to users or groups."),
	"orchestrator sessions":                 newDescriptionInfo("Instance of robot execution", "Sessions represent an instance of a robot's execution of a process."),
	"orchestrator settings":                 newDescriptionInfo("Operation configurations", "Settings define the configurations for various aspects of Orchestrator's operation."),
	"orchestrator stats":                    newDescriptionInfo("Performance metrics", "Stats provide analytics and metrics about the performance and usage of robots and processes."),
	"orchestrator status":                   newDescriptionInfo("Current orchestrator status", "Status represents the current state or progress of various components like jobs, robots, or queue items."),
	"orchestrator task-activities":          newDescriptionInfo("Actions comprising a task", "Task activities are the individual actions or steps that make up a task workflow."),
	"orchestrator task-catalogs":            newDescriptionInfo("Collections of business task definitions", "Task catalogs are organized collections of task definitions for various business processes."),
	"orchestrator task-definitions":         newDescriptionInfo("Settings of a task", "Task definitions describe the properties and settings of a task."),
	"orchestrator task-forms":               newDescriptionInfo("UI for human-robot collaboration tasks", "Task forms are custom UI interfaces created for human-robot collaboration tasks."),
	"orchestrator task-notes":               newDescriptionInfo("Additional details for tasks", "Task notes are additional context or details added to tasks."),
	"orchestrator task-retention":           newDescriptionInfo("Rules for retaining task data", "Task retention is a set of rules that control how long task data is kept before being deleted."),
	"orchestrator tasks":                    newDescriptionInfo("Actions to be completed", "Tasks are actions to be completed in a business process, they may be attended (requires human interaction) or unattended."),
	"orchestrator tenants":                  newDescriptionInfo("Tenants to model your organization structure", "Tenants are independent spaces where you can isolate data, with its own robots, processes and users."),
	"orchestrator test-automation":          newDescriptionInfo("Management and execution of tests", "Test automation is a feature that allows you to plan, manage, and run your automated tests."),
	"orchestrator test-case-definitions":    newDescriptionInfo("Specifications for running test cases", "Test case definitions are the specifications for how a test case is to be run."),
	"orchestrator test-case-executions":     newDescriptionInfo("Instances of tests running", "Test case executions are instances of a test case running on a robot."),
	"orchestrator test-data-queue-actions":  newDescriptionInfo("Operations for managing test data", "Test data queue actions provide operations for managing data in test data queues."),
	"orchestrator test-data-queue-items":    newDescriptionInfo("Units of test data", "Test data queue items are individual units of test data stored in a test data queue."),
	"orchestrator test-data-queues":         newDescriptionInfo("Storage for test data", "Test data queues are used to store test data for consumption in test cases."),
	"orchestrator test-set-executions":      newDescriptionInfo("Instances of test sets running", "Test set executions are instances of a set of test cases running on a robot."),
	"orchestrator test-set-schedules":       newDescriptionInfo("Rules for automatic test runs", "Test set schedules define when a test set is automatically run on a robot."),
	"orchestrator test-sets":                newDescriptionInfo("Groups of test cases", "Test sets are groups of test cases that are set to run together."),
	"orchestrator translations":             newDescriptionInfo("Translation resources", "Translations resources to provide multilingual UI."),
	"orchestrator users":                    newDescriptionInfo("Manage users", "Users are individuals who have access to the Orchestrator's features."),
	"orchestrator webhooks":                 newDescriptionInfo("Notifications of changes", "Webhooks enable real-time notifications about changes or updates in Orchestrator to other applications."),
}

func LookupDescription(name string) *descriptionInfo {
	if result, found := descriptions[name]; found {
		return result
	}
	return nil
}

type descriptionInfo struct {
	Summary     string
	Description string
}

func newDescriptionInfo(summary string, description string) *descriptionInfo {
	return &descriptionInfo{summary, description}
}
