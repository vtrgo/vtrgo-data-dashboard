# Agent Mode Protocol

This document outlines the operational protocol for the agent. The agent will strictly adhere to the modes and rules defined below to ensure a clear, structured, and user-driven workflow.

### **Core Rule: Mode Declaration**

The agent must explicitly state its current operational mode at the very end of every response to the user. The format must be: `Mode: {mode}`.

### **1. Listen Mode**

**Purpose:** To gather high-level project requirements and context directly from the user. In this mode, the user is the sole driver of the project's direction.

**Permissions:**

* The agent is **only** permitted to write to the root-level `GEMINI.md` file.

**Rules of Engagement:**

* The agent's primary role is to listen to the user and update this `GEMINI.md` file with the provided instructions.
* The agent **will not** make any suggestions, offer unsolicited advice, or propose solutions.
* The agent **will never** recommend transitioning to **Execute Mode** directly from **Listen Mode**. The planning phase is a mandatory intermediate step.
* The agent will remain in **Listen Mode** until the user signals that the high-level requirements are sufficiently captured and it's time to move to planning.

### **2. Plan Mode**

**Purpose:** To collaboratively develop a detailed, actionable project plan based on the high-level requirements.

**Permissions:**

* The agent is permitted to create, read, modify, and delete `GEMINI.md` files in **any** directory within the project scope. This allows for folder-specific instructions.

**Rules of Engagement:**

* The agent's role is to ask clarifying questions and gather the specific details required for execution. This includes, but is not limited to, technology stacks, code languages, library choices, file structures, and example shell commands.
* The agent will populate the project's `GEMINI.md` files with this detailed plan.
* The agent is permitted to suggest moving to **Execute Mode** once it assesses that the plan is comprehensive enough to be executed successfully.

### **3. Execute Mode**

**Purpose:** To carry out the detailed project plan.

**Permissions:**

* The agent has full permission to perform any necessary task. This includes creating, modifying, or deleting files and directories, and running shell commands.

**Rules of Engagement:**

* The agent will autonomously execute the tasks outlined in all `GEMINI.md` files throughout the project, respecting both the root-level and folder-specific instructions.
* The agent will provide updates on its progress, successes, and any issues encountered.
* The agent can transition back to **Plan Mode** if the plan proves to be insufficient or requires significant changes based on execution outcomes.

# Project Requirements

- init
- Read the project root level GEMINI.md and go to listen mode.

# Project-Wide `GEMINI.md` Registry

This section serves as a central registry for all `GEMINI.md` files within the project.

**Instructions for the Agent:**
*   During **Plan Mode**, you are responsible for keeping this registry up-to-date.
*   When a new `GEMINI.md` file is created in a subdirectory, you must add a link to it in the list below.
*   When a `GEMINI.md` file is deleted, you must remove its corresponding link from the list.

## Registered Files

*   `/GEMINI.md`
*   `/console/GEMINI.md`
*   `/service/GEMINI.md`