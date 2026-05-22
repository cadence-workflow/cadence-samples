<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Sample Generator

This folder is NOT part of the actual sample. It exists only for contributors who work on this sample. Please disregard it if you are trying to learn about Cadence.

To create a better learning experience for Cadence users, each sample folder is designed to be self contained. Users can view every part of writing and running workflows, including:

* Cadence client initialization
* Worker with workflow and activity registrations
* Workflow starter
* and the workflow code itself

Some samples may have more or fewer parts depending on what they need to demonstrate.

In most cases, the workflow code (e.g. `workflow.go`) is the part that users care about. The rest is boilerplate needed to run that workflow. For each sample folder, the workflow code should be written by hand. The boilerplate can be generated. Keeping all parts inside one folder gives early learners more value because they can see everything together rather than jumping across directories.

## Contributing

* When creating a new sample, follow the steps mentioned in the README file in the main samples folder.
* To update the sample workflow code, edit the workflow file directly.
* To update the worker, client, or other boilerplate logic, edit the generator file. If your change applies to all samples, update the common generator file inside the `template` folder. Edit the generator file in this folder only when the change should affect this sample alone.
* When you are done run the following command in the generator folder

```bash
go run .
```
