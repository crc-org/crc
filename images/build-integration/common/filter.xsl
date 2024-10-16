<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
<xsl:output method="xml" version="1.0" encoding="UTF-8" indent="yes"/>
<xsl:strip-space elements="*"/>

<xsl:template match="@*|node()">
    <xsl:copy>
        <xsl:apply-templates select="@*|node()"/>
    </xsl:copy>
</xsl:template>

<xsl:template match="failure">
    <xsl:variable name="max" select="1000000" />
    <xsl:variable name="length" select="string-length(@message)" />
    <xsl:choose>
        <xsl:when test="$length > $max">
            <xsl:variable name="first" select="substring(@message, 0, floor($max div 2))" />
            <xsl:variable name="second" select="substring(@message, $length - floor($max div 2), $length)" />
            <xsl:copy>
                <xsl:value-of select="concat($first, '\n...\n', $second)"/>
            </xsl:copy>
        </xsl:when>
        <xsl:otherwise>
            <xsl:copy>
                <xsl:value-of select="@message"/>
            </xsl:copy>
        </xsl:otherwise>
    </xsl:choose>
</xsl:template>

</xsl:stylesheet>
