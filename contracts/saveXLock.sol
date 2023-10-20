pragma solidity >=0.7.0 <0.9.0;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract SaveXLock  {
    struct Payment {
        uint256 value;
        address payable receiver;
        bool withdrawn;
        bool cancelled;
    }

    IERC20 public usdtToken;

    mapping(address => Payment[]) public payments;

    event PaymentReceived(address indexed sender, address indexed receiver, uint256 value);
    event PaymentWithdrawn(address indexed receiver, uint256 value);
    event PaymentCancelled(address indexed sender, uint256 value);

    constructor(address _usdtTokenAddress) {
        usdtToken = IERC20(_usdtTokenAddress);
    }         


    function createPayment(address payable _receiver, uint256 _value)  external  {

        require(usdtToken.transferFrom(msg.sender,address(this),_value),"transfer Failed");

        Payment memory newPayment;
        newPayment.value = _value;
        newPayment.receiver = _receiver;
        newPayment.withdrawn = false;
        newPayment.cancelled = false;

        payments[msg.sender].push(newPayment);

        emit PaymentReceived(msg.sender, _receiver, _value);
    }

    function withdrawPayment(uint256 _paymentIndex) external {
        require(_paymentIndex < payments[msg.sender].length, "Invalid payment index");
        Payment storage payment = payments[msg.sender][_paymentIndex];

        require(!payment.withdrawn, "Payment already withdrawn");
        require(!payment.cancelled, "Payment is cancelled");

        payment.withdrawn = true;
        usdtToken.transfer(payment.receiver, payment.value);

        emit PaymentWithdrawn(msg.sender, payment.value);
    }

    function cancelPayment(uint256 _paymentIndex) external {
        require(_paymentIndex < payments[msg.sender].length, "Invalid payment index");
        Payment storage payment = payments[msg.sender][_paymentIndex];

        require(!payment.withdrawn, "Payment already withdrawn");
        require(!payment.cancelled, "Payment is already cancelled");

        payment.cancelled = true;
        usdtToken.transfer(msg.sender, payment.value);

        emit PaymentCancelled(msg.sender, payment.value);
    }
}