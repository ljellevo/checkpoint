## Summary
I want to create a new CLI program called "checkpoint" which works together with GIT to create a checkpoint of the current state of the repository.
This program will be written in Go and will be designed to be fast and efficient, allowing developers to quickly create checkpoints without interrupting their workflow.

It will take the diff of the changed files and store this in the .git folder. Consecutive checkpoints will be stored as a linked list, allowing developers to easily navigate through their checkpoints and revert to previous states if necessary.

They will interact by running "checkpoint" in the terminal, which will create a new checkpoint with the current state of the repository. It's optional to create a message for the checkpoint. This is done by running "checkpoint -m 'message'". 
The checkpoint will be stored in the .git folder, and the user will receive a confirmation message with the checkpoint ID and timestamp.
Typing "checkpoint restore" will restore the repository to the state of the most recent checkpoint, while "checkpoint list" will display a list of all checkpoints with their corresponding messsage, IDs and timestamps.

The changed files will still be visible in the git status, but the checkpoint will allow developers to easily revert to a previous state without having to manually stash or commit their changes. This will be especially useful for developers who want to experiment with new features or make changes without worrying about losing their work.



## Example
1. I change two files (A and B), and add a third file (C). The two changed are being tracked, the new one is not (by git).
2. I run "checkpoint -m "Initial checkpoint".
3. I make a change on file B
4. I run "checkpoint -m "Second checkpoint".
5. I make another change on file B
6. I run "checkpoint restore". The last change on file B is discarded, and im returned to step 4. The latest change on file B is now the one from step 3.
7. I can now run "checkpoint restore" again to return to step 2, where the change on file B from step 3 is also discarded, and I'm back to the state of step 2. The latest change on file B is now the one from step 1, and the change on file A and the new file C are still there, as they were in step 1.
8. As long as i have not made any changes i can run "checkpoint list" and see the two checkpoints i have created. If i at this point make a change the "Second checkpoint" (step 4) will be discarded since the "timeline has changed". If i revert back then the "Second checkpoint" will be back, since the "timeline is back to the state where it was created". If i create a new checkpoint at this point, the "Second checkpoint" will be discarded again, since the "timeline has changed". The "Initial checkpoint" will still be there, since it was created before any of the changes were made.



## Considerations
Navigating back and forth between checkpoints should be easy and intuitive, allowing developers to quickly switch between different states of their repository. 
If a developer navigates back more than 1 checkpoint, then they can not navigate back to the more recent checkpoint if they make a change

The program should also be designed to be fast and efficient, minimizing the time it takes to create and restore checkpoints.


Please use bubbletea (https://github.com/charmbracelet/bubbletea) as the TUI framework for the CLI interface, and create a README.md file with instructions on how to use the program, and how to run it as a developer working on the codebase

## Constraints
- The program must be written in Go.
- The program must interact with GIT to create checkpoints of the current state of the repository.
- The program must store checkpoints in the .git folder as a linked list.
- The program must allow developers to easily navigate through their checkpoints and revert to previous states if necessary.
- The program must be fast and efficient, minimizing the time it takes to create and restore checkpoints.
- The program must use bubbletea as the TUI framework for the CLI interface.
- The "checkpoint list" command should show a vertical timeline somewhat like this:

Start
|
o - Initial checkpoint (#1, 2026-06-11 13:30)
|
o - Second checkpoint (#2, 2026-06-11 13:35)
|
And the user should be able to navigate up and down the timeline using the arrow keys, and select a checkpoint to restore by pressing enter.

## Clarifications
