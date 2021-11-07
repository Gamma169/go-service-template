const chai = require('chai');
const chaiHttp = require('chai-http');
chai.use(chaiHttp);

// Make sure you require the health check endpoint test so that it runs first
require('./health-check-test');


const {
  SERVICE_URL
} = require('./testcases.js');

describe('foobar Tests:', function() {
  before('Setup Database Connections', function(done) {
    setTimeout(done, 250);
  });

  after('Shutdown Database Connections', function(done) {
    setTimeout(done, 250);
  });

  describe('foobar Endpoint Tests', function() {
    
    beforeEach('Populate Database', function(done) {
      setTimeout(done, 250);
    });

    afterEach('Teardown Database', function(done) {
      setTimeout(done, 250);
    });

    describe('Get Model Tests', function() {
      
      // TODO
      it('should be able to get model', function(done) {
        chai.request(SERVICE_URL)
        .get('/health')
        .end(function(err, res){
          if (res) {
            chai.expect(res.status).to.equal(200);
            done();
          }
          else {
            done(err);
          }
        });
      });

    });


    describe('Post Model Tests', function() {
    });


    describe('Update Model Tests', function() {
    });


    describe('Delete Model Tests', function() {
    });


    // End of Endpoints Tests describe
  });
  // End of Test Suite describe
});
