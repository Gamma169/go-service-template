const chai = require('chai');
const chaiHttp = require('chai-http');
chai.use(chaiHttp);

const {SERVICE_URL} = require("./testcases");

describe("Health Check", function() {
  it("should return a 200 code on health endpoint", function(done) {
    chai.request(SERVICE_URL)
      .get('/health')
      .end(function(err, res){
        if (res) {
          chai.expect(res.status).to.equal(200);
          done();
        }
        else {
           console.error("Connection Refused:\n\tPlease check service is running and that service is runnning on " + SERVICE_URL + " which is is open and exposed\n\tTests will not run until a connection has been established"); // eslint-disable-line
          process.exit(1);
        }
      });
  });
});


