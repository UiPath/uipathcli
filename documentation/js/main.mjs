import { Template } from "./template.mjs";

function findCommand(command, name) {
    return command.subcommands.find(c => c.name === name);
}

function render(command, url) {
    let serviceName = null;
    let serviceCommand = null;
    let category = null;
    let categoryCommand = null;
    let operation = null;
    let operationCommand = null;

    const args = url.split('/');
    if (args.length >= 2) {
        serviceName = args[1];
        serviceCommand = findCommand(command, serviceName); 
    }
    if (args.length >= 3 && serviceCommand != null) {
        category = args[2];
        categoryCommand = findCommand(serviceCommand, category); 
    }
    if (args.length >= 4 && categoryCommand != null) {
        operation = args[3];
        operationCommand = findCommand(categoryCommand, operation); 
    }

    const template = new Template();
    if (operationCommand != null) {
        return template.operation(command.name, serviceName, category, operationCommand);
    }
    if (categoryCommand != null) {
        return template.category(command.name, serviceName, categoryCommand);
    }
    if (serviceCommand != null) {
        return template.service(command.name, serviceCommand);
    } 
    return template.main(command);
}

export async function main() {
    const response = await fetch("commands.json");
    const command = await response.json();
    const element = document.querySelector('.main');

    window.onhashchange = function() {
        const template = render(command, window.location.hash);
        element.innerHTML = template;
    };
    const template = render(command, window.location.hash);
    element.innerHTML = template;
}