export default
  window.location.port === '3000'
    ? "//" + window.location.hostname + ":8080/api"
    : "//" + window.location.host + "/api";
