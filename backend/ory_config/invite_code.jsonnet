function (ctx) {
    user_id: ctx.identity.id,
    invitation_code: ctx.flow.transient_payload.invitation_code
}