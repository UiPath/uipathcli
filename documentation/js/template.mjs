export class Template {
    main(command) {
        return `
            <h1>${command.name}</h1>
            <h2>Description</h2>
            <div class="description">${command.description}</div>
            <h2>Available Services</h2>
            <ul class="services">
                ${command.subcommands.map(subcommand => `
                <li><a href="#/${subcommand.name}">${subcommand.name}</a></li>
                `).join('')}
            </ul>
            <h2>Configuration</h2>
            <div class="configuration">
                The CLI supports multiple ways to authorize with the UiPath services:
                <ul>
                    <li><b>Client Credentials</b>: Generate secret and configure the CLI to use these long-term credentials. Client credentials should be used in case you want to use the CLI from a script in an automated way.</li>
                    <li><b>OAuth Login</b>: Login to UiPath using your browser and SSO of choice. This is the preferred flow when you are using the CLI interactively. No need to manage any credentials.</li>
                </ul>
            </div>
            <h2>Client Credentials</h2>
            <div class="configuration">
                <p>
                    In order to use client credentials, you need to set up an <a href="https://docs.uipath.com/automation-cloud/docs/managing-external-applications">External Application (Confidential)</a> and generate an <a href="https://docs.uipath.com/automation-suite/docs/managing-external-applications#generating-a-new-app-secret">application secret</a>.
                </p>
                <p>Run the interactive CLI configuration:</p>
                <code>uipath config --auth credentials</code>
                <p>
                    The CLI will ask you to enter the main config settings like
                    <ul>
                        <li>organization and tenant used by UiPath services which are account-scoped or tenant-scoped</li>
                        <li>clientId and clientSecret to retrieve the JWT bearer token for authentication</li>
                    </ul>
                </p>
                <p>After that the CLI should be ready and you can validate that it is working by invoking one of the services (requires OR.Users.Read scope):</p>
                <code>uipath orchestrator users get</code>
            </div>
            <h2>OAuth Login</h2>
            <div class="configuration">
                <p>
                    In order to use oauth login, you need to set up an <a href="https://docs.uipath.com/automation-cloud/automation-cloud/latest/admin-guide/managing-external-applications">External Application (Non-Confidential)</a> with a redirect url which points to your local CLI:
                </p>
                <p>Run the interactive CLI configuration:</p>
                <code>uipath config --auth login</code>
                <p>
                    The CLI will ask you to enter the main config settings like
                    <ul>
                        <li>organization and tenant used by UiPath services which are account-scoped or tenant-scoped</li>
                        <li>clientId, redirectUri and scopes which are needed to initiate the OAuth flow</li>
                    </ul>
                </p>
                <p>After that the CLI should be ready and you can validate that it is working by invoking one of the services:</p>
                <code>uipath orchestrator users get</code>
            </div>
            <h2>Global Parameters</h2>
            <ul class="parameters">
                ${command.parameters.map(parameter => `
                <li class="parameter">
                    <span class="parameter-name">--${parameter.name} ${parameter.type}</span>
                    <span class="parameter-required">${parameter.required ? `(required)` : ''}</span>
                    <p class="parameter-description">${parameter.description}</p>
                    <div class="parameter-default-value">${parameter.defaultValue ? `
                        <span>Default Value: ${parameter.defaultValue}</span>` : ''}
                    </div>
                    <div class="parameter-allowed-values">${parameter.allowedValues && parameter.allowedValues.length > 0 ? 
                    `<span>Allowed Values:</span>
                        <ul>
                            ${parameter.allowedValues.map(value => `
                            <li>${value}</li>
                            `).join('')}
                        </ul>` : ''}
                    </div>
                    <div class="parameter-example">${parameter.example ? `
                        <span>Example:</span><br/>
                        <p class="parameter-example-code">"${parameter.example}"</p>` : ''}
                    </div>
                </li>
                `).join('')}
            </ul>
        `;
    }
    
    service(executableName, command) {
        return `
            <nav class="breadcrumb">
                <ol>
                    <li><a href="#/">${executableName}</a></li>
                    <li>${command.name}</li>
                </ol>
            </nav>
            <h1>${executableName} ${command.name}</h1>
            <h2>Description</h2>
            <div class="description">${command.description}</div>
            <h2>Available Commands</h2>
            <ul class="commands">
                ${command.subcommands.map(subcommand => `
                <li><a href="#/${command.name}/${subcommand.name}">${subcommand.name}</a></li>
                `).join('')}
            </ul>
        `;
    }

    category(executableName, serviceName, command) {
        return `
            <nav class="breadcrumb">
                <ol>
                    <li><a href="#/">${executableName}</a></li>
                    <li><a href="#/${serviceName}">${serviceName}</a></li>
                    <li>${command.name}</li>
                </ol>
            </nav>
            <h1>${executableName} ${serviceName} ${command.name}</h1>
            <h2>Description</h2>
            <div class="description">${command.description}</div>
            <h2>Available Commands</h2>
            <ul class="commands">
                ${command.subcommands.map(subcommand => `
                <li><a href="#/${serviceName}/${command.name}/${subcommand.name}">${subcommand.name}</a></li>
                `).join('')}
            </ul>
        `;
    }

    operation(executableName, serviceName, categoryName, command) {
       return `
            <nav class="breadcrumb">
                <ol>
                    <li><a href="#/">${executableName}</a></li>
                    <li><a href="#/${serviceName}">${serviceName}</a></li>
                    <li><a href="#/${serviceName}/${categoryName}">${categoryName}</a></li>
                    <li>${command.name}</li>
                </ol>
            </nav>
            <h1>${executableName} ${serviceName} ${categoryName} ${command.name}</h1>
            <h2>Description</h2>
            <div class="description">${command.description}</div>
            <h2>Usage</h2>
            <div class="usage">
                ${executableName} ${serviceName} ${categoryName} ${command.name}<br/>
                ${command.parameters.map(parameter => `
                &nbsp;&nbsp;${parameter.required ? '' : '['}--${parameter.name} ${parameter.type}${parameter.required ? '' : ']'}<br/>
                `).join('')}
            </div>
            <h2>Parameters</h2>
            <ul class="parameters">
                ${command.parameters.map(parameter => `
                <li class="parameter">
                    <span class="parameter-name">--${parameter.name} ${parameter.type}</span>
                    <span class="parameter-required">${parameter.required ? `(required)` : ''}</span>
                    <p class="parameter-description">${parameter.description}</p>
                    <div class="parameter-default-value">${parameter.defaultValue ? `
                        <span>Default Value: ${parameter.defaultValue}</span>` : ''}
                    </div>
                    <div class="parameter-allowed-values">${parameter.allowedValues && parameter.allowedValues.length > 0 ? 
                    `<span>Allowed Values:</span>
                        <ul>
                            ${parameter.allowedValues.map(value => `
                            <li>${value}</li>
                            `).join('')}
                        </ul>` : ''}
                    </div>
                    <div class="parameter-example">${parameter.example ? `
                        <span>Example:</span><br/>
                        <p class="parameter-example-code">${parameter.example}</p>` : ''}
                    </div>
                </li>
                `).join('')}
            </ul>
        `;
    }
}