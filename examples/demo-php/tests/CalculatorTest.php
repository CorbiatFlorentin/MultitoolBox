<?php

declare(strict_types=1);

namespace Demo\Tests;

use Demo\Calculator;
use PHPUnit\Framework\TestCase;

final class CalculatorTest extends TestCase
{
    public function testAdd(): void
    {
        $this->assertSame(5, (new Calculator())->add(2, 3));
    }

    public function testDivideByZeroThrows(): void
    {
        $this->expectException(\DivisionByZeroError::class);
        (new Calculator())->divide(1, 0);
    }
}
