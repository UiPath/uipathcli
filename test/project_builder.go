package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const ProjectJsonTemplate = `
{
  "name": "%s",
  "projectId": "%s",
  "description": "Blank Process",
  "main": "Main.xaml",
  "dependencies": {
    "UiPath.System.Activities": "[24.10.3]",
    "UiPath.Testing.Activities": "[24.10.0]",
    "UiPath.UIAutomation.Activities": "[24.10.0]"
  },
  "webServices": [],
  "entitiesStores": [],
  "schemaVersion": "4.0",
  "studioVersion": "24.10.1.0",
  "projectVersion": "1.0.0",
  "runtimeOptions": {
    "autoDispose": false,
    "netFrameworkLazyLoading": false,
    "isPausable": true,
    "isAttended": false,
    "requiresUserInteraction": false,
    "supportsPersistence": false,
    "workflowSerialization": "DataContract",
    "excludedLoggedData": [
      "Private:*",
      "*password*"
    ],
    "executionType": "Workflow",
    "readyForPiP": false,
    "startsInPiP": false,
    "mustRestoreAllDependencies": true,
    "pipType": "ChildSession",
    "robotVersion": "25.0.0"
  },
  "designOptions": {
    "projectProfile": "Developement",
    "outputType": "Process",
    "libraryOptions": {
      "includeOriginalXaml": false,
      "privateWorkflows": []
    },
    "processOptions": {
      "ignoredFiles": [
        "MyProcess.*.nupkg"
      ]
    },
    "fileInfoCollection": [],
    "saveToCloud": false
  },
  "expressionLanguage": "CSharp",
  "entryPoints": [
    {
      "filePath": "Main.xaml",
      "uniqueId": "ac610120-f85b-4ed4-a014-05dca3380186",
      "input": [],
      "output": []
    }
  ],
  "isTemplate": false,
  "templateProjectData": {},
  "publishData": {},
  "targetFramework": "%s"
}
`

const MainXaml = `
<Activity mc:Ignorable="sap sap2010" x:Class="Main" sap2010:ExpressionActivityEditor.ExpressionActivityEditor="C#" sap:VirtualizedContainerService.HintSize="1088.8,636.8" sap2010:WorkflowViewState.IdRef="ActivityBuilder_1" xmlns="http://schemas.microsoft.com/netfx/2009/xaml/activities" xmlns:mc="http://schemas.openxmlformats.org/markup-compatibility/2006" xmlns:sap="http://schemas.microsoft.com/netfx/2009/xaml/activities/presentation" xmlns:sap2010="http://schemas.microsoft.com/netfx/2010/xaml/activities/presentation" xmlns:scg="clr-namespace:System.Collections.Generic;assembly=System.Private.CoreLib" xmlns:sco="clr-namespace:System.Collections.ObjectModel;assembly=System.Private.CoreLib" xmlns:ui="http://schemas.uipath.com/workflow/activities" xmlns:x="http://schemas.microsoft.com/winfx/2006/xaml">
  <TextExpression.NamespacesForImplementation>
    <sco:Collection x:TypeArguments="x:String">
      <x:String>System.Activities</x:String>
      <x:String>System.Activities.Statements</x:String>
      <x:String>System.Activities.Expressions</x:String>
      <x:String>System.Activities.Validation</x:String>
      <x:String>System.Activities.XamlIntegration</x:String>
      <x:String>Microsoft.VisualBasic</x:String>
      <x:String>Microsoft.VisualBasic.Activities</x:String>
      <x:String>System</x:String>
      <x:String>System.Collections</x:String>
      <x:String>System.Collections.Generic</x:String>
      <x:String>System.Collections.ObjectModel</x:String>
      <x:String>System.Data</x:String>
      <x:String>System.Diagnostics</x:String>
      <x:String>System.Drawing</x:String>
      <x:String>System.IO</x:String>
      <x:String>System.Linq</x:String>
      <x:String>System.Net.Mail</x:String>
      <x:String>System.Xml</x:String>
      <x:String>System.Text</x:String>
      <x:String>System.Xml.Linq</x:String>
      <x:String>UiPath.Core</x:String>
      <x:String>UiPath.Core.Activities</x:String>
      <x:String>System.Windows.Markup</x:String>
      <x:String>GlobalVariablesNamespace</x:String>
      <x:String>GlobalConstantsNamespace</x:String>
      <x:String>System.Linq.Expressions</x:String>
      <x:String>System.Runtime.Serialization</x:String>
    </sco:Collection>
  </TextExpression.NamespacesForImplementation>
  <TextExpression.ReferencesForImplementation>
    <sco:Collection x:TypeArguments="AssemblyReference">
      <AssemblyReference>Microsoft.CSharp</AssemblyReference>
      <AssemblyReference>Microsoft.VisualBasic</AssemblyReference>
      <AssemblyReference>mscorlib</AssemblyReference>
      <AssemblyReference>System</AssemblyReference>
      <AssemblyReference>System.Activities</AssemblyReference>
      <AssemblyReference>System.ComponentModel.TypeConverter</AssemblyReference>
      <AssemblyReference>System.Core</AssemblyReference>
      <AssemblyReference>System.Data</AssemblyReference>
      <AssemblyReference>System.Data.Common</AssemblyReference>
      <AssemblyReference>System.Data.DataSetExtensions</AssemblyReference>
      <AssemblyReference>System.Drawing</AssemblyReference>
      <AssemblyReference>System.Drawing.Common</AssemblyReference>
      <AssemblyReference>System.Drawing.Primitives</AssemblyReference>
      <AssemblyReference>System.Linq</AssemblyReference>
      <AssemblyReference>System.Net.Mail</AssemblyReference>
      <AssemblyReference>System.ObjectModel</AssemblyReference>
      <AssemblyReference>System.Private.CoreLib</AssemblyReference>
      <AssemblyReference>System.Runtime.Serialization</AssemblyReference>
      <AssemblyReference>System.ServiceModel</AssemblyReference>
      <AssemblyReference>System.ServiceModel.Activities</AssemblyReference>
      <AssemblyReference>System.Xaml</AssemblyReference>
      <AssemblyReference>System.Xml</AssemblyReference>
      <AssemblyReference>System.Xml.Linq</AssemblyReference>
      <AssemblyReference>UiPath.System.Activities</AssemblyReference>
      <AssemblyReference>UiPath.UiAutomation.Activities</AssemblyReference>
      <AssemblyReference>UiPath.Studio.Constants</AssemblyReference>
      <AssemblyReference>System.Console</AssemblyReference>
      <AssemblyReference>System.Security.Permissions</AssemblyReference>
      <AssemblyReference>System.Configuration.ConfigurationManager</AssemblyReference>
      <AssemblyReference>System.ComponentModel</AssemblyReference>
      <AssemblyReference>System.Memory</AssemblyReference>
      <AssemblyReference>System.Private.Uri</AssemblyReference>
      <AssemblyReference>System.Linq.Expressions</AssemblyReference>
      <AssemblyReference>System.Runtime.Serialization.Formatters</AssemblyReference>
      <AssemblyReference>System.Private.DataContractSerialization</AssemblyReference>
      <AssemblyReference>System.Runtime.Serialization.Primitives</AssemblyReference>
    </sco:Collection>
  </TextExpression.ReferencesForImplementation>
  <Sequence DisplayName="Main Sequence" sap:VirtualizedContainerService.HintSize="416,254.4" sap2010:WorkflowViewState.IdRef="Sequence_1">
    <sap:WorkflowViewStateService.ViewState>
      <scg:Dictionary x:TypeArguments="x:String, x:Object">
        <x:Boolean x:Key="IsExpanded">True</x:Boolean>
      </scg:Dictionary>
    </sap:WorkflowViewStateService.ViewState>
    <ui:LogMessage DisplayName="Log Message" sap:VirtualizedContainerService.HintSize="353.6,165.6" sap2010:WorkflowViewState.IdRef="LogMessage_1">
      <ui:LogMessage.Message>
        <InArgument x:TypeArguments="x:Object">
          <CSharpValue x:TypeArguments="x:Object">"Hello World"</CSharpValue>
        </InArgument>
      </ui:LogMessage.Message>
    </ui:LogMessage>
  </Sequence>
</Activity>
`

const DefaultGovernanceFile = `
{
  "product-name": "Development",
  "policy-name": "Modern Policy - Development",
  "data": {
    "core-allow-edit": true,
    "experimental-features-dictionary": {},
    "telemetry-redirection-options": {
      "instrumentation-keys": ""
    },
    "publish-source-control-info": null,
    "enforce-checkin-before-publish": false,
    "enforce-checkin-before-publish-allow-edit": true,
    "enforce-repositories-config-allow-edit": true,
    "enforce-repositories-config": {
      "enforce-repositories-config-allow-save-local": null,
      "enforce-repositories-config-allow-edit-repositories": null,
      "enforce-repositories-config-repositories": [
        {
          "enforce-repositories-config-repository-default-folder": null,
          "enforce-repositories-config-repository-name": null,
          "enforce-repositories-config-repository-url": null,
          "enforce-repositories-config-repository-source-control-type": "Git"
        }
      ]
    },
    "compiler-optimizations": false,
    "compiler-optimizations-allow-edit": true,
    "hide-starting-screen": false,
    "feedback-enabled": true,
    "require-user-publish-permitted-consecutive-runs": "",
    "require-user-publish-dialog-message": null,
    "require-user-publish-log-to-queue-queue-name": null,
    "require-user-publish-log-to-queue-queue-folder": null,
    "template-feeds": [
      {
        "template-feed-name": "GettingStarted",
        "template-feed-is-enabled": true
      },
      {
        "template-feed-name": "Official",
        "template-feed-is-enabled": true
      },
      {
        "template-feed-name": "Marketplace",
        "template-feed-is-enabled": true
      }
    ],
    "user-add-feeds": false,
    "user-enable-feeds": false,
    "append-orchestrator-feeds": true,
    "preview-packages-enabled": null,
    "preview-packages-enabled-allow-edit": true,
    "default-package-source": [
      {
        "default-package-source-name": "nuget.org",
        "default-package-source-source": "https://api.nuget.org/v3/index.json",
        "is-enabled-default-package-source": true
      }
    ],
    "send-ui-descriptors": false,
    "send-ui-descriptors-allow-edit": false,
    "create-docked-annotations": true,
    "create-docked-annotations-allow-edit": true,
    "enforce-analyzer-before-run": false,
    "enforce-analyzer-before-run-allow-edit": true,
    "use-smart-file-paths": true,
    "use-smart-file-paths-allow-edit": true,
    "enable-activity-online-recommendations": true,
    "enable-activity-online-recommendations-allow-edit": true,
    "enable-discovered-activities": true,
    "enable-discovered-activities-allow-edit": true,
    "enforce-release-notes": false,
    "display-legacy-framework-deprecation": true,
    "enforce-analyzer-before-publish": false,
    "enforce-analyzer-before-publish-allow-edit": true,
    "publish-applications-metadata": true,
    "enforce-analyzer-before-push": false,
    "enforce-analyzer-before-push-allow-edit": true,
    "is-collapsed-view-slim": true,
    "is-collapsed-view-slim-allow-edit": true,
    "analyze-rpa-xamls-only": false,
    "analyze-rpa-xamls-only-allow-edit": true,
    "object-repository-enforced": false,
    "object-repository-enforced-allow-edit": true,
    "default-project-language": "VisualBasic",
    "default-project-language-allow-edit": true,
    "default-project-framework": "Modern",
    "default-project-framework-allow-edit": true,
    "allowed-project-frameworks": {
      "Classic": false,
      "Modern": true,
      "CrossPlatform": true
    },
    "additional-analyzer-rule-path": null,
    "additional-analyzer-rule-path-allow-edit": true,
    "project-path": null,
    "project-path-allow-edit": true,
    "publish-process-url": null,
    "publish-process-url-allow-edit": true,
    "publish-library-url": null,
    "publish-library-url-allow-edit": true,
    "publish-templates-url": null,
    "publish-templates-url-allow-edit": true,
    "export-analyzer-results": false,
    "export-analyzer-results-allow-edit": true,
    "use-connection-service": false,
    "use-connection-service-allow-edit": true,
    "default-pip-type": "ChildSession",
    "default-pip-type-allow-edit": true,
    "show-data-manager-only": false,
    "show-data-manager-only-allow-edit": true,
    "show-properties-inline": false,
    "show-properties-inline-allow-edit": true,
    "allowed-publish-feeds": {
      "Custom": true,
      "PersonalWorkspace": true,
      "TenantPackages": true,
      "FolderPackages": true,
      "HostLibraries": true,
      "TenantLibraries": true,
      "Local": true
    },
    "auto-generate-activity-outputs": true,
    "auto-generate-activity-outputs-allow-edit": true,
    "separate-runtime-dependencies": true,
    "separate-runtime-dependencies-allow-edit": true,
    "include-sources": true,
    "include-sources-allow-edit": true,
    "save-to-cloud-by-default": null,
    "save-to-cloud-by-default-allow-edit": true,
    "enable-generative-ai": false,
    "activities-manager-allow-show-developer": true,
    "activities-manager-hidden-activities": [],
    "analyzer-allow-edit": false,
    "referenced-rules-config-file": null,
    "embedded-rules-config-rules": [
      {
        "code-embedded-rules-config-rules": "ST-ANA-006",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-ANA-005",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-USG-034",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-USG-032",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "RequiredTags",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "ST-USG-027",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "RequiredPackages",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "ST-USG-014",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "ProhibitedPackages",
            "value-use-default-value-rules": true
          },
          {
            "key-embedded-rules-config-rules": "AllowedPackages",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "TA-DBP-003",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Info",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-SEC-008",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "VariableDepthUsage",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "ST-SEC-007",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-SEC-009",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "Excluded",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "ST-PRR-004",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-USG-028",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-USG-026",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "ProhibitedActivities",
            "value-use-default-value-rules": true
          },
          {
            "key-embedded-rules-config-rules": "AllowedActivities",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "ST-USG-009",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-MRD-009",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "Threshold",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "ST-MRD-011",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-MRD-005",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-MRD-008",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-MRD-004",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-MRD-007",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "Threshold",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "ST-DBP-026",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-DBP-024",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-DBP-028",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-DBP-025",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-DBP-023",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-DBP-007",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "LayersCount",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "ST-DBP-002",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "ArgumentsCount",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "ST-NMG-016",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "Length",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "ST-NMG-006",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-NMG-012",
        "is-enabled-embedded-rules-config-rules": false,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-NMG-011",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-NMG-002",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "InRegex",
            "value-use-default-value-rules": true
          },
          {
            "key-embedded-rules-config-rules": "OutRegex",
            "value-use-default-value-rules": true
          },
          {
            "key-embedded-rules-config-rules": "InOutRegex",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "ST-NMG-005",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-NMG-004",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "Threshold",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "ST-DBP-027",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-USG-020",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "TA-DBP-007",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "BranchesMaxCount",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "TA-DBP-006",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "ActivitiesMaxCount",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "TA-NMG-002",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "RegEx",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "TA-DBP-001",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "VerificationsMinCount",
            "value-use-default-value-rules": true
          },
          {
            "key-embedded-rules-config-rules": "VerificationsMaxCount",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "TA-NMG-001",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-USG-010",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "TA-DBP-004",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "TA-DBP-002",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "UI-PST-001",
        "is-enabled-embedded-rules-config-rules": false,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "logLevel",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "UI-ANA-016",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Info",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "UI-ANA-017",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Info",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-MRD-017",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-REL-006",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-USG-005",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "IncludeActivities",
            "value-use-default-value-rules": true
          },
          {
            "key-embedded-rules-config-rules": "ExcludeActivities",
            "value-use-default-value-rules": true
          },
          {
            "key-embedded-rules-config-rules": "IncludeProperties",
            "value-use-default-value-rules": true
          },
          {
            "key-embedded-rules-config-rules": "ExcludeProperties",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "ST-MRD-002",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-DBP-020",
        "is-enabled-embedded-rules-config-rules": false,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "Excluded",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "ST-DBP-003",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-NMG-009",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "ST-NMG-008",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "Length",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "ST-NMG-001",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "Regex",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "SY-USG-013",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "UI-USG-011",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "nonAllowedAttributes",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "UX-SEC-010",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "blacklistApps",
            "value-use-default-value-rules": true
          },
          {
            "key-embedded-rules-config-rules": "whitelistApps",
            "value-use-default-value-rules": true
          },
          {
            "key-embedded-rules-config-rules": "blacklistUrls",
            "value-use-default-value-rules": true
          },
          {
            "key-embedded-rules-config-rules": "whitelistUrls",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "UI-SEC-004",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "UI-REL-001",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "idxValue",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "UI-PRR-004",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "UX-DBP-029",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "UI-DBP-030",
        "is-enabled-embedded-rules-config-rules": false,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "UI-DBP-013",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "UI-ANA-018",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Info",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "MA-DBP-028",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "UI-SEC-010",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": [
          {
            "key-embedded-rules-config-rules": "blacklistApps",
            "value-use-default-value-rules": true
          },
          {
            "key-embedded-rules-config-rules": "whitelistApps",
            "value-use-default-value-rules": true
          },
          {
            "key-embedded-rules-config-rules": "blacklistUrls",
            "value-use-default-value-rules": true
          },
          {
            "key-embedded-rules-config-rules": "whitelistUrls",
            "value-use-default-value-rules": true
          }
        ]
      },
      {
        "code-embedded-rules-config-rules": "ST-DBP-021",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "UI-PRR-003",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "UI-PRR-002",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "UI-PRR-001",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "XL-DBP-027",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "SY-USG-014",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "UI-DBP-006",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": []
      },
      {
        "code-embedded-rules-config-rules": "TA-DBP-005",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      }
    ],
    "embedded-rules-config-counter": [
      {
        "code-embedded-rules-config-counter": "ST-ANA-003",
        "is-enabled-embedded-rules-config-counter": true,
        "parameters-embedded-rules-config-counter": []
      },
      {
        "code-embedded-rules-config-counter": "ST-ANA-009",
        "is-enabled-embedded-rules-config-counter": true,
        "parameters-embedded-rules-config-counter": []
      }
    ]
  }
}
`

type ProjectBuilder struct {
	t                     *testing.T
	projectName           string
	projectId             string
	targetFramework       string
	defaultGovernanceFile string
}

func (b ProjectBuilder) buildProjectJson() string {
	return fmt.Sprintf(ProjectJsonTemplate, b.projectName, b.projectId, b.targetFramework)
}

func (b *ProjectBuilder) WithProjectName(name string) *ProjectBuilder {
	b.projectName = name
	return b
}

func (b *ProjectBuilder) WithDefaultGovernanceFile() *ProjectBuilder {
	b.defaultGovernanceFile = DefaultGovernanceFile
	return b
}

func (b ProjectBuilder) writeFileContent(directory string, fileName string, content string) {
	err := os.WriteFile(filepath.Join(directory, fileName), []byte(content), 0600)
	if err != nil {
		b.t.Fatal(err)
	}
}

func (b ProjectBuilder) Build() string {
	directory := CreateDirectory(b.t)

	projectJson := b.buildProjectJson()
	b.writeFileContent(directory, "project.json", projectJson)
	b.writeFileContent(directory, "Main.xaml", MainXaml)
	if b.defaultGovernanceFile != "" {
		b.writeFileContent(directory, "uipath.policy.default.json", b.defaultGovernanceFile)
	}

	return directory
}

func NewCrossPlatformProject(t *testing.T) *ProjectBuilder {
	return &ProjectBuilder{
		t,
		"MyProcess",
		"9011ee47-8dd4-4726-8850-299bd6ef057c",
		"Portable",
		"",
	}
}

func NewWindowsProject(t *testing.T) *ProjectBuilder {
	return &ProjectBuilder{
		t,
		"MyWindowsProcess",
		"94c4c9c1-68c3-45d4-be14-d4427f17eddd",
		"Windows",
		"",
	}
}
