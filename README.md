## CloudKeep Backend

### OVERVIEW

"CloudKeep" is a cloud-based file storage and sharing platform. Aims to provide a features similar to google drive service.

### QUICK START

Download & Install Docker
`git clone https://github.com/shripadaRao/CloudKeep.git `
`sudo docker-compose up --build`

### API DESIGN

##### User Registration

Description: Allows users to create a new account. To configure the template of email, head to config/registerEmailTemplate.json.

Endpoints:

1. `/api/register/send-email-otp` `POST`
2. `/api/register/verify-otp` `POST`
3. `/api/register/create-user` `POST`

### System Design

##### User Registrations

<img src="assets/user_registrations_workflow.png" alt="drawing" width="800" height="1000"/>
