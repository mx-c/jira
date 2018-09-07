# Warning

This is a POC and not even close to a companywide usable software.

## Rationale

For some operations, it is really much more handy to remain in the terminal env.
In particular, creating hooks within the git process to check up if there are notifications, update ticket status upon related branch creation ...

## 
1.create an api basic token on :https://id.atlassian.com
2. create the ~/.jira.json file that must hold the following:
{
    "email": "yourEmailRegisteredOnJira@mention.com",
    "basic_token": "api_token_created",
    "cloud_id": "used-for-notifs",
    "cloud_session_token": "session_token_retrieved_from_navigation" 
}

> currently, we use a basic authentication workflow, defined [here](https://confluence.atlassian.com/cloud/api-tokens-938839638.html), when possible.


## Limitations
currently, we use the session_token retrieved from an open session. These sessions are long living it seems (duration 1 month). in a second step, we should login and request with the dedicated Client.

### Definition
    
    transition
    When you switch the status of a ticket.
    
    edition
    Update ticket's attribute.

// notifications
