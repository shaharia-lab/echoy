# This dockerfile is only used to build the backend image for the application using goreleaser.
FROM scratch
COPY echoy /app/echoy
ENTRYPOINT []